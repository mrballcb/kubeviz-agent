package data

import (
  "time"
  "github.com/bartlettc22/kubeviz-agent/pkg/kubernetes"
  "github.com/bartlettc22/kubeviz-agent/pkg/aws"
)

type DataStruct struct {
  Metadata Metadata
  KubernetesResources kubernetes.KubernetesResources
  AwsResources aws.AwsResources
  // ResourceList []*metav1.APIResourceList
  // Nodes []v1.Node
  // Namespaces []v1.Namespace
}

type Metadata struct {
  // ClusterName string
  // KubernetesVersion string
  // AwsAccount string
  // AwsAccountAlias string
  AgentVersion string
  RunTime time.Time
  RunDuration time.Duration
}

var Data DataStruct
