import = "aws-sdk-go/models/apis/cloud9/2017-09-23/api-2.json"

mapping "aws_cloud9_environment_ec2" {
  name                        = EnvironmentName
  instance_type               = InstanceType
  automatic_stop_time_minutes = AutomaticStopTimeMinutes
  description                 = EnvironmentDescription
  owner_arn                   = UserArn
  subnet_id                   = SubnetId
}

test "aws_cloud9_environment_ec2" "instance_type" {
  ok = "t2.micro"
  ng = "t20.micro"
}

test "aws_cloud9_environment_ec2" "owner_arn" {
  ok = "arn:aws:iam::123456789012:user/David"
  ng = "arn:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment"
}
