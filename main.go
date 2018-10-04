package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "os"
  "path/filepath"
  v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
  "k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
  restclient "k8s.io/client-go/rest"
  log "github.com/Sirupsen/logrus"
)

type Info struct {
  Version string `json:"version"`
  Nodes []v1.Node `json:"nodes"`
}

var info Info
var config *restclient.Config
var clientset *kubernetes.Clientset
var err error

func main() {

  var kubeconfig *string

  log.SetLevel(log.DebugLevel)

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

  getVersion()
  getNodes()

  output, err := json.Marshal(info)
  if err != nil {
    log.Fatal("Unable to create JSON output", err)
  }

  fmt.Println(string(output))

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getVersion() {
  client := discovery.NewDiscoveryClientForConfigOrDie(config)
  version, err := client.ServerVersion()
  if err != nil {
    log.Fatal("Error fetching Kubernetes version")
  }

  info.Version = version.GitVersion
}

func getNodes() {
  nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}
  info.Nodes = nodes.Items
}
