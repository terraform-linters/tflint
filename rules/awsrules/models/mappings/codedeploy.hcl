import = "aws-sdk-go/models/apis/codedeploy/2014-10-06/api-2.json"

mapping "aws_codedeploy_app" {
  name             = ApplicationName
  compute_platform = ComputePlatform
}

mapping "aws_codedeploy_deployment_config" {
  deployment_config_name = DeploymentConfigName
  compute_platform       = ComputePlatform
}

mapping "aws_codedeploy_deployment_group" {
  app_name               = ApplicationName
  deployment_group_name  = DeploymentGroupName
  deployment_config_name = DeploymentConfigName
}

test "aws_codedeploy_app" "compute_platform" {
  ok = "Server"
  ng = "Fargate"
}
