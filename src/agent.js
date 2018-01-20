var AWS = require('aws-sdk');
var Promise = require('bluebird');
var fs = Promise.promisifyAll(require('fs'));

var ec2 = new AWS.EC2();
var autoscaling = new AWS.AutoScaling();
var elb = new AWS.ELB();
var elbv2 = new AWS.ELBv2();
var route53 = new AWS.Route53();
var iam = new AWS.IAM();
var sts = new AWS.STS();

var lbArns = []
var lbDns = []

main();

function main() {

  AWS.config.setPromisesDependency(require('bluebird'));

  // Store the provided hosted zone ids to search for matches
  var instanceIds = process.argv[2].split("\n");
  var hostedZoneIds = process.argv[3].split(",");

  // Write empty file first
  //fs.writeFileAsync("data.json", "[]")
  //.then(function() {
  getInstances(instanceIds)
  .then(function(asgNames) {
    return getAutoscalingGroups(asgNames);
  }).then(function(targetGroupsAndLoadBalancers) {
    // Returns a combined list of v1 and v2 load balancer dns names
    return getTargetGroups(targetGroupsAndLoadBalancers[0])
      .then(function (loadBalancerArns) {
        lbArns = loadBalancerArns
        return getV2LoadBalancers(loadBalancerArns)
          .then(function(loadBalancerArns) {
            return getLoadBalancerListeners(loadBalancerArns);
          }).then(function(listenerArns) {
            return getListenRules(listenerArns);
          })
      })
      .then(function () {
        return getV1LoadBalancers(targetGroupsAndLoadBalancers[1]);
      })
  }).then(function(loadBalancerDnsNames) {
    return getMatchingDnsRecords(loadBalancerDnsNames, hostedZoneIds);
  }).then(function() {
    return getAccountInfo();
  });
}

function getInstances(InstanceIds) {
  var params = {
    InstanceIds: InstanceIds
  }
  return ec2.describeInstances(params).promise()
  .then(function(data) {
    reservations = data.Reservations;           // successful response
    var instanceData = [];
    var AsgNames = [];
    for(i=0;i<reservations.length;i++) {
      for(j=0;j<reservations[i].Instances.length; j++) {
        reservations[i].Instances[j]["kind"] = "Instance"
        instanceData.push(reservations[i].Instances[j]);
        for(k=0;k<reservations[i].Instances[j].Tags.length; k++) {
          if(reservations[i].Instances[j].Tags[k].Key === "aws:autoscaling:groupName") {
            var id = reservations[i].Instances[j].Tags[k].Value
            if(AsgNames.indexOf(id) === -1){
              AsgNames.push(id);
            }
          }
        }
      }
    }

    return writeResource("instance",instanceData).then(function() { return AsgNames})
  })
}

function getAutoscalingGroups(AsgNames) {
  var params = {
    AutoScalingGroupNames: AsgNames
  };
  return autoscaling.describeAutoScalingGroups(params).promise()
  .then(function(data) {
    var autoscalingGroupData = [];
    var targetGroupARNs = [];
    var loadBalancerNames = []
    asgs=data.AutoScalingGroups
    for(i=0;i<asgs.length;i++) {

      // Add unique target groups to array
      for(j in asgs[i].TargetGroupARNs) {
        if(targetGroupARNs.indexOf(asgs[i].TargetGroupARNs[j]) === -1){
          targetGroupARNs.push(asgs[i].TargetGroupARNs[j]);
        }
      }

      // Add unique load balancer names to array
      for(j in asgs[i].LoadBalancerNames) {
        if(loadBalancerNames.indexOf(asgs[i].LoadBalancerNames[j]) === -1){
          loadBalancerNames.push(asgs[i].LoadBalancerNames[j]);
        }
      }

      asgs[i]["kind"] = "AutoscalingGroup"
      autoscalingGroupData.push(asgs[i]);
    }
    return writeResource("asg",autoscalingGroupData).then(function() { return [targetGroupARNs,loadBalancerNames]})
  })
}

function getTargetGroups(targetGroupARNs) {
  var params = {
    TargetGroupArns: targetGroupARNs
  };
  return elbv2.describeTargetGroups(params).promise()
  .then(function(data) {
    var targetGroupData = [];
    var loadBalancerArns = []
    tgs=data.TargetGroups
    for(i=0;i<tgs.length;i++) {

      // Add unique load balancer names to array
      for(j in tgs[i].LoadBalancerArns) {
        if(loadBalancerArns.indexOf(tgs[i].LoadBalancerArns[j]) === -1){
          loadBalancerArns.push(tgs[i].LoadBalancerArns[j]);
        }
      }

      tgs[i]["kind"] = "TargetGroup"
      targetGroupData.push(tgs[i]);
    }
    return writeResource("tg",targetGroupData).then(function() { return loadBalancerArns})
  })
}

function getV2LoadBalancers(loadBalancerArns) {
  var params = {
    LoadBalancerArns: loadBalancerArns
  };
  return elbv2.describeLoadBalancers(params).promise()
  .then(function(data) {
    var loadBalancerData = [];
    // var loadBalancerDnsNames = [];
    var loadBalancerArns = [];
    lbs=data.LoadBalancers
    for(i=0;i<lbs.length;i++) {
      lbDns.push(lbs[i].DNSName)
      loadBalancerArns.push(lbs[i].LoadBalancerArn)
      lbs[i]["kind"] = "LoadBalancerV2"
      loadBalancerData.push(lbs[i]);
    }
    return writeResource("lbv2",loadBalancerData).then(function() { return loadBalancerArns })
  })
}

function getLoadBalancerListeners(loadBalancerArns) {
  var promises = [];
  for(i=0;i<loadBalancerArns.length;i++) {
    var params = {
      LoadBalancerArn: loadBalancerArns[i]
    };
    promises.push(
      elbv2.describeListeners(params).promise()
        .then(function(data) {
          // var listenerData = [];
          var listenerArns = [];
          listeners=data.Listeners
          for(j=0;j<listeners.length;j++) {
            listenerArns.push(listeners[j].ListenerArn)
            listeners[j]["kind"] = "LoadBalancerListener"
            // listenerData.push(listeners[j])
          }
          return writeResource("lblistener",listeners).then(function() { return listenerArns })
        })
    )
  }
  return Promise.all(promises)
}

// Fetches ALB Listener Rules
// Input 2-d array of listener arns [[arn1, arn2],[arn3,arn4]]
function getListenRules(listenerArns) {
  promises = []

  // listenerArns comes in as a 2-d array
  for(i=0;i<listenerArns.length;i++) {
    for(j=0;j<listenerArns[i].length;j++) {
      var params ={
        ListenerArn: listenerArns[i][j]
      };
      promises.push(
        elbv2.describeRules(params).promise()
        .then(function(data) {
          var ruleArns = []
          for(k=0;k<data.Rules.length;k++) {
            ruleArns.push(data.Rules[k].RuleArn);
            data.Rules[k]["kind"] = "ListenerRule"
          }
          return writeResource("listenerrule",data.Rules).then(function() { return ruleArns })
        })
      )
    }
  }

  return Promise.all(promises)
}

function getV1LoadBalancers(loadBalancerNames) {
  var params = {
    LoadBalancerNames: loadBalancerNames
  };
  return elb.describeLoadBalancers(params).promise()
  .then(function(data) {
    loadBalancerData = [];
    lbs=data.LoadBalancerDescriptions
    for(i=0;i<lbs.length;i++) {
      lbDns.push(lbs[i].DNSName)
      lbs[i]["kind"] = "LoadBalancer"
      loadBalancerData.push(lbs[i]);
    }
    return writeResource("lb",loadBalancerData).then(function() { return lbDns})
  })
}

function getMatchingDnsRecords(loadBalancerDnsNames, hostedZoneIds) {
  var promises = [];

  for (i in hostedZoneIds) {
    var params = {
      Id: hostedZoneIds[i]
    };
    promises.push(
      route53.getHostedZone(params).promise()
      .then(function(hostedZoneData) {
        var hZData = hostedZoneData
        var HostedZoneId = hostedZoneData.HostedZone.Id.split("/")[2];
        hZData["HostedZoneId"] = HostedZoneId
        hZData["kind"] = "HostedZone"
        return writeResource("hz",hZData).then(function() { return HostedZoneId })
      })
      .then(function(HostedZoneId) {
        var params = {
          HostedZoneId:HostedZoneId
        };
        return route53.listResourceRecordSets(params).promise()
        .then(function(data) {
          var dnsRecordData = [];
          var dnsRecordNames = []
          var records=data.ResourceRecordSets
          for(j=0;j<records.length;j++) {
            if(records[j].Type == "CNAME") {
              for(k=0;k<records[j].ResourceRecords.length;k++) {
                if(loadBalancerDnsNames.indexOf(records[j].ResourceRecords[k].Value) !== -1){
                  dnsRecordNames.push(records[j].Name);
                  records[j]["kind"] = "DnsRecord"
                  records[j]["hostedZoneId"] = hostedZoneIds[i]
                  dnsRecordData.push(records[j]);
                }
              }
            } else if(records[j].Type == "A" && records[j].AliasTarget) {
              var rawdnsname=records[j].AliasTarget.DNSName
              var dnsname = rawdnsname.substring(0, rawdnsname.length - 1)
              // console.log("recording alias record", dnsname)
              if(loadBalancerDnsNames.indexOf(dnsname) !== -1){
                dnsRecordNames.push(records[j].Name);
                records[j]["kind"] = "DnsRecord"
                records[j]["DNSName"] = dnsname
                records[j]["hostedZoneId"] = hostedZoneIds[i]
                dnsRecordData.push(records[j]);
              }
            }
          }
          return writeResource("dns",dnsRecordData).then(function() { return dnsRecordNames })
        })
      })
    )
  }

  return Promise.all(promises)
}

function getAccountInfo() {
  return iam.listAccountAliases({}).promise()
  .then(function(data) {
    accountData = {}
    accountData.AccountAliases = data.AccountAliases
    accountData.kind = "awsAccount"
    return accountData
  }).then(function(accountData) {
    return sts.getCallerIdentity({}).promise()
    .then(function(data) {
      accountData.Account = data.Account
      return writeResource("aa",accountData)
    })
  })
}

function writeResource(type, resource_data) {

  return fs.readFileAsync('/data/data.json')
  .then(function(data) {

    // Parse the existing data in the file

    try {
      var json = JSON.parse(data)
    } catch(e) {

      console.log(e); // error in the above string (in this case, yes)!
      console.log(type)
    }

    // Push new data into the array
    json = json.concat(resource_data)

    // Write the new json to file
    return fs.writeFileAsync("/data/data.json", JSON.stringify(json));
  })
}
