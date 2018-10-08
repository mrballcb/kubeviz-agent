package kubernetes

import (
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

type KubernetesResources struct {
  Metadata Metadata
  // ResourceList []*metav1.APIResourceList
  Nodes []v1.Node
  Namespaces []v1.Namespace
  Pods []v1.Pod
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

func init() {

  home := homeDir();
  kubeConfigPath := filepath.Join(home, ".kube", "config")
  if e, _ := exists(kubeConfigPath); home != "" && e {
    log.Info("Attempting .kube/config")
    // use the current context in kubeconfig
    kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
    if err != nil {
      log.Fatal(err.Error())
    }
  } else {
    // Try in-cluster config
    log.Info("Attempting in-cluster config")
    kubeConfig, err = restclient.InClusterConfig()
  	if err != nil {
  		log.Fatal(err.Error())
  	}
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
  getPods(resources)
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

func getPods(resources *KubernetesResources) {
  pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}
  resources.Pods = pods.Items
}

func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}
