# Kubernetes Visualization Helm Chart

## Configuration
| Parameter               | Description                           | Default                                                    |
| ----------------------- | ----------------------------------    | ---------------------------------------------------------- |
| `Agent.ComponentName` | Used for resource names and labeling | `kubeviz-agent` |
| `Agent.Image` | Kubeviz agent image | `bartlettc/kubeviz-agent` |
| `Agent.ImageTag` | Container image tag | `0.1.0` |
| `Agent.ImagePullPolicy` | Container image pull policy | `IfNotPresent` |
| `Agent.ApiEndpoint` | Kubeviz Server endpoint | ` ` |
| `Agent.ApiKey` | Kubeviz Server API key | ` ` |
| `Agent.Interval` | Interval for querying K8s/AWS | `60` |
| `Agent.AwsAccessKey` | AWS Access Key for account Kubernetes is running in | ` ` |
| `Agent.AwsSecretKey` | AWS Secret Key for account Kubernetes is running in | ` ` |
| `rbac.create` | Create rbac roles (set to true if rbac is enabled) | `false` |
