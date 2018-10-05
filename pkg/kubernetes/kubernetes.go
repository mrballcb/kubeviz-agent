package kubernetes

import (
  // "encoding/json"
  "flag"
  // "fmt"
  "os"
  "path/filepath"
  v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
  "k8s.io/client-go/discovery"
  // "k8s.io/client-go/discovery/helper"
	"k8s.io/client-go/tools/clientcmd"
  restclient "k8s.io/client-go/rest"
  log "github.com/Sirupsen/logrus"
)

type KubernetesResources struct {
  Metadata Metadata
  // ResourceList []*metav1.APIResourceList
  Nodes []v1.Node
  Namespaces []v1.Namespace
}

type Metadata struct {
  ClusterName string
  KubernetesVersion string
  Region string
}

var Resources KubernetesResources
var kubeConfig *restclient.Config
var clientset *kubernetes.Clientset
var err error

func Init() {

  var kubeconfig *string

  if home := homeDir(); home != "" {
    kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
  } else {
    kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
  }
  flag.Parse()

  // use the current context in kubeconfig
  kubeConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
  if err != nil {
    panic(err.Error())
  }

  clientset, err = kubernetes.NewForConfig(kubeConfig)
  if err != nil {
    log.Fatal(err.Error())
  }
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func Run(resources *KubernetesResources) {
  getVersion(resources)
  getNodes(resources)
  getNamespaces(resources)
}

func getVersion(resources *KubernetesResources) {
  client := discovery.NewDiscoveryClientForConfigOrDie(kubeConfig)
  version, err := client.ServerVersion()
  if err != nil {
    log.Fatal("Error fetching Kubernetes version")
  }

  resources.Metadata.KubernetesVersion = version.GitVersion
}

func getNodes(resources *KubernetesResources) {
  nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}

  // Since cluster name is not really a thing in Kubernetes, we use the custom
  // node label that we've added to every cluster
  resources.Metadata.ClusterName = nodes.Items[0].ObjectMeta.Labels["clusterName"]
  resources.Metadata.Region = nodes.Items[0].ObjectMeta.Labels["failure-domain.beta.kubernetes.io/region"]
  resources.Nodes = nodes.Items
}

func getNamespaces(resources *KubernetesResources) {
  namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}
  resources.Namespaces = namespaces.Items
}

// func getServices() {
//   namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
//   cluster.Namespaces = namespaces.Items
// }
//
// func getIngresses() {
//   namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
//   cluster.Namespaces = namespaces.Items
// }
