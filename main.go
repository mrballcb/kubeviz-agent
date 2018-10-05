package main

import (
  // "encoding/json"
  // "flag"
  // "fmt"
  // "os"
  "encoding/json"
  // "path/filepath"
  "time"
  // v1 "k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/kubernetes"
  // "k8s.io/client-go/discovery"
  // "k8s.io/client-go/discovery/helper"
	// "k8s.io/client-go/tools/clientcmd"
  // restclient "k8s.io/client-go/rest"
  log "github.com/Sirupsen/logrus"
  // envconfig "github.com/kelseyhightower/envconfig"
  // "github.com/spf13/viper"
  "github.com/bartlettc22/kubeviz-agent/pkg/kubernetes"
  "github.com/bartlettc22/kubeviz-agent/pkg/aws"
  "github.com/bartlettc22/kubeviz-agent/pkg/data"
)



// var config EnvConfig
// var data Data

var err error

// type EnvConfig struct {
//
// }

func main() {

  log.SetLevel(log.DebugLevel)

  data.Data.Metadata.AgentVersion = "0.2.0"

  // Initialize kubernetes configuration
  // Will use cluster api if inside Kubernetes
  // Will use kubeconfig if outside Kubernetes
  kubernetes.Init()


  aws.Init()

  tick := time.Tick(time.Duration(10000) * time.Millisecond)
  run()
  for range tick {
    log.Debug("Starting new run")
    run()
  }

  // client := discovery.NewDiscoveryClientForConfigOrDie(config)
  // info.ResourceList, err = client.ServerResources()
  // if err != nil {
  //   log.Fatal(err)
  // }
  //
  // fmt.Println(discovery.GroupVersionResources(info.ResourceList))
  //
  //
  //
  // // x, _ := discovery.GroupVersionResources(info.ResourceList)
  // // output, err := json.Marshal(x)
  // // if err != nil {
  // //   log.Fatal("Unable to create JSON output", err)
  // // }
  // //
  // // fmt.Println(string(output))
  //





  //
  // fmt.Println(string(output))

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

  post(data)
}
