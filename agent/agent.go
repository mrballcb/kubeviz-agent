package agent

import (
  "encoding/json"
  "time"
  log "github.com/Sirupsen/logrus"
  "github.com/bartlettc22/kubeviz-agent/pkg/kubernetes"
  "github.com/bartlettc22/kubeviz-agent/pkg/aws"
  "github.com/bartlettc22/kubeviz-agent/pkg/data"
)

var serverAddress, tokenAuth string

func Start(address string, token string) {

  serverAddress = address
  tokenAuth = token

  data.Data.Metadata.AgentVersion = "0.2.0"

  // Initialize kubernetes configuration
  // Will use cluster api if inside Kubernetes
  // Will use kubeconfig if outside Kubernetes
  kubernetes.Init()
  aws.Init()
  log.Info("test")
  tick := time.Tick(time.Duration(10000) * time.Millisecond)
  run()
  for range tick {
    log.Debug("Starting new run")
    run()
  }
}

func run() {
  start := time.Now()
  data.Data.Metadata.RunTime = start

  kubernetes.Run(&data.Data.KubernetesResources)
  aws.Run(&data.Data.AwsResources, &kubernetes.Resources.Metadata.ClusterName)

  data.Data.Metadata.RunDuration = time.Since(start)

  data, err := json.Marshal(data.Data)
  if err != nil {
    log.Fatal("Unable to create JSON output", err)
  }

  post(data, serverAddress, tokenAuth)
}
