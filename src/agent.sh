#!/bin/bash

echo "[INFO] Agent (v${AGENT_VERSION}) Started"

# Ensure ENVs are set
if [[ ! -v API_ENDPOINT ]]; then
    echo "[ERROR] API_ENDPOINT is required to be set"
    exit -1
fi

if [[ ! -v API_KEY ]]; then
    echo "[ERROR] API_KEY is required to be set"
    exit -1
fi

if [[ ! -v DNS_ZONE_IDS ]]; then
    echo "[ERROR] DNS_ZONE_IDS is required to be set"
    exit -1
fi

if [[ ! -v KOPS_STATE_STORE ]]; then
    echo "[ERROR] KOPS_STATE_STORE is required to be set"
    exit -1
fi

json_escape () {
  JSON_TOPIC_RAW=$1
  JSON_TOPIC_RAW=${JSON_TOPIC_RAW//\\/\\\\} # \
  JSON_TOPIC_RAW=${JSON_TOPIC_RAW//\//\\\/} # /
  JSON_TOPIC_RAW=${JSON_TOPIC_RAW//\"/\\\"} # "
  JSON_TOPIC_RAW=${JSON_TOPIC_RAW//   /\\t} # \t (tab)
  JSON_TOPIC_RAW=${JSON_TOPIC_RAW//
/\\\n} # \n (newline)
  JSON_TOPIC_RAW=${JSON_TOPIC_RAW//^M/\\\r} # \r (carriage return)
  JSON_TOPIC_RAW=${JSON_TOPIC_RAW//^L/\\\f} # \f (form feed)
  JSON_TOPIC_RAW=${JSON_TOPIC_RAW//^H/\\\b} # \b (backspace)
  echo $JSON_TOPIC_RAW
}

vercomp () {
    if [[ $1 == $2 ]]
    then
        return 0
    fi
    local IFS=.
    local i ver1=($1) ver2=($2)
    # fill empty fields in ver1 with zeros
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++))
    do
        ver1[i]=0
    done
    for ((i=0; i<${#ver1[@]}; i++))
    do
        if [[ -z ${ver2[i]} ]]
        then
            # fill empty fields in ver2 with zeros
            ver2[i]=0
        fi
        if ((10#${ver1[i]} > 10#${ver2[i]}))
        then
            return 1
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]}))
        then
            return 2
        fi
    done
    return 0
}

while true;

  do

  echo "[INFO] Agent (v${AGENT_VERSION}) Run Started"

  START_DATE=$(date)
  START_SECONDS=$SECONDS
  echo "" > /data/data.json
  echo "" > /data/error.log

  if [ -z "$REMOTE_RUN" ]; then
    EC2_AVAIL_ZONE=`curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone`
    export AWS_REGION="`echo \"$EC2_AVAIL_ZONE\" | sed -e 's:\([0-9][0-9]*\)[a-z]*\$:\\1:'`"
  fi

  # Get the cluster name that we're in
  # Assumes ALL the nodes are labeld with 'clusterName'
  # Another way to get cluster?: $(kubectl get po -l k8s-app=kube-apiserver -n kube-system -o json | jq -r '.items[0].metadata.annotations["dns.alpha.kubernetes.io/external"]' | sed 's/^api\.//')
  CLUSTER_NAME=$(kubectl get no -o json | jq -r '.items[0].metadata.labels.clusterName' 2>> /data/error.log)
  CLUSTER_VERSION=$(kops get cluster --name $CLUSTER_NAME -o json | jq -r .spec.kubernetesVersion 2>> /data/error.log)
  echo "[INFO] Cluster: $CLUSTER_NAME"

  echo "[INFO] Collecing K8S resources..."
  # Bring it all together
  ADDITIONAL_RESOURCES=""
  vercomp $CLUSTER_VERSION "1.8.0"
  if [ $? -eq 0 -o $? -eq 1 ]; then
    ADDITIONAL_RESOURCES=",cronjobs"
  fi
  kubectl get deploy,ds,rs,statefulset,po,no,ns,ing,svc,endpoints,jobs${ADDITIONAL_RESOURCES} --all-namespaces -o json | \
  jq '.items[] |= del(.spec?.template?.spec?.containers[]?.env) | .items' | \
  jq '.[] |= del(.metadata?.annotations["kubectl.kubernetes.io/last-applied-configuration"]?)' | \
  jq '.[] |= del(.spec?.containers[]?.env)' | \
  jq '. |= . + ['"$(kops get cluster --name $CLUSTER_NAME -o json 2>> /data/error.log)"']' | \
  jq '. |= . + '"$(kops get ig --name $CLUSTER_NAME -o json | jq '.[] |= . + {"cluster":"'$CLUSTER_NAME'"}')" \
  >> /data/data.json 2>> /data/error.log

  # echo "[]" > /data/data.json

  # externalID = AWS Instance Ids
  CLUSTER_NODES=$(kubectl get no -o json | jq -r '.items[].spec.externalID' 2>> /data/error.log)

  echo "[INFO] Collecing AWS resources (matching against DNS Zones $DNS_ZONE_IDS)..."
  nodejs agent.js "$CLUSTER_NODES" "$DNS_ZONE_IDS" >> /data/error.log 2>&1
  NODE_SUCCESS=$?

  ELAPSED_TIME=$(($SECONDS - $START_SECONDS))
  STATS='{"kind":"agentStats","agent_version":"'${AGENT_VERSION}'","cluster_version":"'${CLUSTER_VERSION}'","query_time":"'${ELAPSED_TIME}'","run_start":"'${START_DATE}'","run_interval":"'${AGENT_INTERVAL}'","errors":"'$(json_escape "$(cat /data/error.log)")'"}'

  cat /data/data.json | jq '. += ['"${STATS}"']' > /data/data_final.json

  echo "[INFO] Sending data to collection server $API_ENDPOINT..."
  if [ -z "$REMOTE_RUN" ]; then
    if [ $NODE_SUCCESS -eq 0 ]; then
      curl \
        -s \
        -H "Content-Type: application/json" \
        -H "X-KubeViz-Token: $API_KEY" \
        -X POST \
        -d @/data/data_final.json \
        $API_ENDPOINT/data?cluster=$CLUSTER_NAME
    else
      echo "AWS fetch failed, not sending..."
    fi
  fi

  echo "[INFO] Cleaning Up"
  if [ -z "$REMOTE_RUN" ]; then
    rm /data/data.json
  fi

  echo "[INFO] Done"

  sleep $AGENT_INTERVAL;
done
