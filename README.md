# Kubernetes Visualization Agent
This agent reports Kubernetes and AWS details back to a central [kubeviz server](https://github.com/bartlettc22/kubeviz-server)

## Helm Installation
```
helm upgrade \
  --install \
  --namespace kubeviz \
  --set Agent.ApiEndpoint=${API_ENDPOINT} \
  --set Agent.ApiKey=${API_KEY} \
  --set Agent.AwsAccessKey=${AWS_ACCESS_KEY_ID} \
  --set Agent.AwsSecretKey=${AWS_SECRET_ACCESS_KEY} \
  --set rbac.create=<true|false> \
  kubeviz-agent ./helm_chart/kubeviz-agent/
```

See [helm_chart/kubeviz-agent](helm_chart/kubeviz-agent) for more information on the Helm configuration.

### AWS
The `AwsAccessKey` and `AwsSecretKey` must have the following access in order for AWS information to be queried.
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "ec2:Describe*",
                "ec2:Get*",
                "iam:Get*",
                "iam:List*",
                "sts:Get*"
            ],
            "Effect": "Allow",
            "Resource": "*"
        }
    ]
}
```
