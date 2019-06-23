import = "aws-sdk-go/models/apis/elasticmapreduce/2009-03-31/api-2.json"

mapping "aws_emr_cluster" {
  name                              = String
  release_label                     = String
  master_instance_group             = any // InstanceGroup
  master_instance_type              = any // InstanceType
  scale_down_behavior               = ScaleDownBehavior
  additional_info                   = any // XmlString
  service_role                      = String
  security_configuration            = any // XmlString
  core_instance_group               = InstanceGroup
  core_instance_type                = any // InstanceType
  core_instance_count               = Integer
  instance_group                    = InstanceGroup
  log_uri                           = String
  applications                      = ApplicationList
  termination_protection            = Boolean
  keep_job_flow_alive_when_no_steps = Boolean
  ec2_attributes                    = Ec2InstanceAttributes
  kerberos_attributes               = KerberosAttributes
  ebs_root_volume_size              = Integer
  custom_ami_id                     = any // XmlStringMaxLen256
  bootstrap_action                  = BootstrapActionConfigList
  configurations                    = ConfigurationList
  configurations_json               = String
  visible_to_all_users              = Boolean
  autoscaling_role                  = any // XmlString
  step                              = StepConfigList
  tags                              = TagList
}

mapping "aws_emr_instance_group" {
  name               = any // XmlStringMaxLen256
  cluster_id         = ClusterId
  instance_type      = any // InstanceType
  instance_count     = Integer
  bid_price          = any // XmlStringMaxLen256
  ebs_optimized      = BooleanObject
  ebs_config         = EbsBlockDeviceConfig
  autoscaling_policy = AutoScalingPolicyDescription
}

mapping "aws_emr_security_configuration" {
  name          = any // XmlString
  name_prefix   = String
  configuration = String
}
