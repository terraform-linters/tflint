import = "aws-sdk-go/models/apis/autoscaling/2011-01-01/api-2.json"

mapping "aws_launch_configuration" {
  name                             = CreateLaunchConfigurationType
  // name_prefix                   = String
  // image_id                      = XmlStringMaxLen255
  // instance_type                 = XmlStringMaxLen255
  // iam_instance_profile          = XmlStringMaxLen1600
  // key_name                      = XmlStringMaxLen255
  security_groups                  = SecurityGroups
  // associate_public_ip_address   = String
  // vpc_classic_link_id           = XmlStringMaxLen255
  vpc_classic_link_security_groups = ClassicLinkVPCSecurityGroups
  // user_data                     = XmlStringUserData
  // user_data_base64              = String
  enable_monitoring                = InstanceMonitoring
  ebs_optimized                    = EbsOptimized
  root_block_device                = BlockDeviceMappings
  ebs_block_device                 = BlockDeviceMappings
  ephemeral_block_device           = BlockDeviceMappings
  spot_price                       = SpotPrice
  // placement_tenancy             = XmlStringMaxLen64
}
