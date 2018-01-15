#!/bin/bash

# Ensure ENVs are set
if [[ ! -v API_ENDPOINT ]]; then
    echo "API_ENDPOINT is required to be set"
    exit -1
fi

if [[ ! -v API_KEY ]]; then
    echo "API_KEY is required to be set"
    exit -1
fi

if [[ ! -v DNS_ZONE_IDS ]]; then
    echo "DNS_ZONE_IDS is required to be set"
    exit -1
fi

if [[ ! -v KOPS_STATE_STORE ]]; then
    echo "KOPS_STATE_STORE is required to be set"
    exit -1
fi

if [ -z "$REMOTE_RUN" ]; then
  EC2_AVAIL_ZONE=`curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone`
  export AWS_REGION="`echo \"$EC2_AVAIL_ZONE\" | sed -e 's:\([0-9][0-9]*\)[a-z]*\$:\\1:'`"
fi

while true;

  do
  # Get the cluster name that we're in
  # Assumes ALL the nodes are labeld with 'clusterName'
  # Another way to get cluster?: $(kubectl get po -l k8s-app=kube-apiserver -n kube-system -o json | jq -r '.items[0].metadata.annotations["dns.alpha.kubernetes.io/external"]' | sed 's/^api\.//')
  CLUSTER_NAME=$(kubectl get no -o json | jq -r '.items[0].metadata.labels.clusterName')
  echo "[INFO] Cluster: $CLUSTER_NAME"

  echo "[INFO] Collecing K8S resources..."

  # Ugly shit to parse out the `kubectl describe svc` command. Please help
  SERVICES=$(kubectl get svc --all-namespaces -o json | jq -r '.items[] | .metadata.name+"/"+.metadata.namespace')
  SERVICE_DESCRIPTIONS=""
  for SVC in $SERVICES;
    do
      SVC_NAME=$(echo "$SVC" | awk -v delimeter="/" '{split($0,a,delimeter)} END{print a[1]}')
      SVC_NS=$(echo "$SVC" | awk -v delimeter="/" '{split($0,a,delimeter)} END{print a[2]}')
      SVC_DESC=$(echo "{"$(kubectl describe svc $SVC_NAME -n $SVC_NS | sed -E '/^( |Labels|Annotations).*$/d' | sed -E 's/^([a-zA-Z ]+):[ ]+(.*)/"\1":"\2",/g')"\"kind\":\"ServiceDescription\",\"cluster\":\"$CLUSTER_NAME\"}")
      SERVICE_DESCRIPTIONS=$(echo "${SERVICE_DESCRIPTIONS}${SVC_DESC},")
    done

  # Remove trailing comma and add brackets
  SERVICE_DESCRIPTIONS="["${SERVICE_DESCRIPTIONS%,*}"]"

  # Bring it all together
  kubectl get deploy,ds,rs,statefulset,po,no,ns,ing,svc --all-namespaces -o json | \
  jq '.items[] |= . + {"cluster":"'$CLUSTER_NAME'"} | .items' | \
  jq '.[] |= del(.spec?.template?.spec?.containers[]?.env)' | \
  jq '.[] |= del(.metadata?.annotations["kubectl.kubernetes.io/last-applied-configuration"]?)' | \
  jq '.[] |= del(.spec?.containers[]?.env)' | \
  jq '. |= . + '"$SERVICE_DESCRIPTIONS" | \
  jq '. |= . + ['"$(kops get cluster --name $CLUSTER_NAME -o json)"']' | \
  jq '. |= . + '"$(kops get ig --name $CLUSTER_NAME -o json | jq '.[] |= . + {"cluster":"'$CLUSTER_NAME'"}')" \
  > /data/data.json

  # echo "[]" > /data/data.json

  # externalID = AWS Instance Ids
  CLUSTER_NODES=$(kubectl get no -o json | jq -r '.items[].spec.externalID')

  echo "[INFO] Collecing AWS resources (matching against DNS Zones $DNS_ZONE_IDS)..."
  echo "[INFO] Instances to be scanned $CLUSTER_NODES"
  nodejs agent.js "$CLUSTER_NODES" "$DNS_ZONE_IDS"
  NODE_SUCCESS=$?

  echo "[INFO] Sending data to collection server $API_ENDPOINT..."
  echo "$NODE_SUCCESS"
  if [ -z "$REMOTE_RUN" ]; then
    if [ $NODE_SUCCESS -eq 0 ]; then
      curl \
        -H "Content-Type: application/json" \
        -H "X-KubeViz-Token: $API_KEY" \
        -X POST \
        -d @/data/data.json \
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

  sleep 60;
done
