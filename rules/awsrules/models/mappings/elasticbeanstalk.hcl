import = "aws-sdk-go/models/apis/elasticbeanstalk/2010-12-01/api-2.json"

mapping "aws_elastic_beanstalk_application" {
  name        = ApplicationName
  description = Description
  tags        = Tags
}

mapping "aws_elastic_beanstalk_application_version" {
  name         = VersionLabel
  application  = ApplicationName
  description  = Description
  bucket       = S3Bucket
  key          = S3Key
  force_delete = ForceTerminate
  tags         = Tags
}

mapping "aws_elastic_beanstalk_configuration_template" {
  name                = ConfigurationTemplateName
  application         = ApplicationName
  description         = Description
  environment_id      = EnvironmentId
  setting             = ConfigurationOptionSettingsList
  solution_stack_name = SolutionStackName
}

mapping "aws_elastic_beanstalk_environment" {
  name                   = EnvironmentName
  application            = ApplicationName
  cname_prefix           = DNSCnamePrefix
  description            = Description
  tier                   = EnvironmentTier
  setting                = ConfigurationOptionSettingsList
  solution_stack_name    = SolutionStackName
  template_name          = ConfigurationTemplateName
  platform_arn           = PlatformArn
  version_label          = VersionLabel
  tags                   = Tags
}
