import = "aws-sdk-go/models/apis/eks/2017-11-01/api-2.json"

mapping "aws_eks_cluster" {
  name                      = ClusterName
  role_arn                  = String
  vpc_config                = VpcConfigRequest
  enabled_cluster_log_types = Logging
  version                   = String
}

test "aws_eks_cluster" "name" {
  ok = "example"
  ng = "@example"
}
