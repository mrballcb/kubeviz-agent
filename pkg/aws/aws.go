package aws

import (
  "strings"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  // "github.com/aws/aws-sdk-go/service/autoscaling"
  // "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
  // "github.com/aws/aws-sdk-go/service/ec2"
  "github.com/aws/aws-sdk-go/service/iam"
  "github.com/aws/aws-sdk-go/service/sts"
  log "github.com/Sirupsen/logrus"
)

type AwsResources struct {
  Metadata Metadata
}

type Metadata struct {
  // ClusterName string
  // KubernetesVersion string
  AwsAccount string
  AwsAccountAlias string

  // AgentVersion string
  // RunTime time.Time
  // RunDuration time.Duration
}

var iamClient *iam.IAM
var stsClient *sts.STS

var Resources AwsResources

// var ec2Client *ec2.EC2
// var autoscalingClient *autoscaling.AutoScaling
// var resourcegroupstaggingapiClient *resourcegroupstaggingapi.ResourceGroupsTaggingAPI

func Init() {
  sess := session.Must(session.NewSession(&aws.Config{
	Region: aws.String("us-west-2"),
}))
  iamClient = iam.New(sess)
  stsClient = sts.New(sess)
  // ec2Client = ec2.New(sess)
  // autoscalingClient = autoscaling.New(sess)
  // resourcegroupstaggingapiClient = resourcegroupstaggingapi.New(sess)
}

func Run(resources *AwsResources, clusterName *string) {
  getAwsAccount(resources)
  getAwsAccountAlias(resources)
  // getAwsAutoscaling(clusterName)
}

func getAwsAccountAlias(resources *AwsResources) {
  result, err := iamClient.ListAccountAliases(&iam.ListAccountAliasesInput{})
  if err != nil {
    log.Warn("Error", err)
    return
  }

  var aliases []string
  for _, alias := range result.AccountAliases {
    if alias == nil {
        continue
    }
    aliases = append(aliases, *alias)
  }

  resources.Metadata.AwsAccountAlias = strings.Join(aliases, "/")
}

func getAwsAccount(resources *AwsResources) {
  result, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
  if err != nil {
    log.Warn("Error", err)
    return
  }

  account := result.Account
  resources.Metadata.AwsAccount = *account
}

// func getAwsAutoscaling(clusterName *string) {
//  log.Info(*clusterName)
//   tagKey := "k8s.io/cluster-autoscaler/node-template/label/clusterName"
//   tagValue := *clusterName
//   tagFilter := resourcegroupstaggingapi.TagFilter{Key: &tagKey, Values: []*string{&tagValue}}
// log.Info(tagFilter)
//   result, err := resourcegroupstaggingapiClient.GetResources(&resourcegroupstaggingapi.GetResourcesInput{
//     TagFilters: []*resourcegroupstaggingapi.TagFilter{&tagFilter},
//   })
//   if err != nil {
//     log.Warn("Error", err)
//     return
//   }
//
//   log.Info(result.ResourceTagMappingList)
//   // account := result.Account
//   // cluster.Metadata.AwsAccount = *account
// }
