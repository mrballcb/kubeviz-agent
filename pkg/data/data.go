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
}

type Metadata struct {
  AgentVersion string
  RunTime time.Time
  RunDuration time.Duration
}

var Data DataStruct
