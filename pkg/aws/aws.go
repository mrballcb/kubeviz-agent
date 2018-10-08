package aws

import (
  "strings"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/iam"
  "github.com/aws/aws-sdk-go/service/sts"
  log "github.com/Sirupsen/logrus"
)

type AwsResources struct {
  Metadata Metadata
}

type Metadata struct {
  AwsAccount string
  AwsAccountAlias string
}

var iamClient *iam.IAM
var stsClient *sts.STS
var Resources AwsResources

func init() {
  sess := session.Must(session.NewSession(&aws.Config{
	   Region: aws.String("us-west-2"),
  }))
  iamClient = iam.New(sess)
  stsClient = sts.New(sess)
}

func Run(resources *AwsResources, clusterName *string) {
  getAwsAccount(resources)
  getAwsAccountAlias(resources)
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
