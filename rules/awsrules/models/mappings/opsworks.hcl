import = "aws-sdk-go/models/apis/opsworks/2013-02-18/api-2.json"

mapping "aws_opsworks_application" {
  name                      = String
  short_name                = String
  stack_id                  = String
  type                      = AppType
  description               = String
  environment               = EnvironmentVariables
  enable_ssl                = Boolean
  ssl_configuration         = SslConfiguration
  app_source                = Source
  data_source_arn           = String
  data_source_type          = String
  data_source_database_name = String
  domains                   = Strings
  document_root             = String
  auto_bundle_on_deploy     = String
  rails_env                 = String
  aws_flow_ruby_settings    = String
}

mapping "aws_opsworks_custom_layer" {
  name                        = String
  short_name                  = String
  stack_id                    = String
  auto_assign_elastic_ips     = Boolean
  auto_assign_public_ips      = Boolean
  custom_instance_profile_arn = String
  custom_security_group_ids   = Strings
  auto_healing                = Boolean
  install_updates_on_boot     = Boolean
  instance_shutdown_timeout   = Integer
  elastic_load_balancer       = String
  drain_elb_on_shutdown       = Boolean
  system_packages             = Strings
  use_ebs_optimized_instances = Boolean
  ebs_volume                  = VolumeConfiguration
  custom_json                 = String
  custom_configure_recipes    = Strings
  custom_deploy_recipes       = Strings
  custom_setup_recipes        = Strings
  custom_shutdown_recipes     = Strings
  custom_undeploy_recipes     = Strings
}

mapping "aws_opsworks_ganglia_layer" {
  stack_id                    = String
  password                    = String
  name                        = String
  auto_assign_elastic_ips     = Boolean
  auto_assign_public_ips      = Boolean
  custom_instance_profile_arn = String
  custom_security_group_ids   = Strings
  auto_healing                = Boolean
  install_updates_on_boot     = Boolean
  instance_shutdown_timeout   = Integer
  elastic_load_balancer       = String
  drain_elb_on_shutdown       = Boolean
  system_packages             = Strings
  url                         = String
  username                    = String
  use_ebs_optimized_instances = Boolean
  ebs_volume                  = VolumeConfiguration
  custom_json                 = String
  custom_configure_recipes    = Strings
  custom_deploy_recipes       = Strings
  custom_setup_recipes        = Strings
  custom_shutdown_recipes     = Strings
  custom_undeploy_recipes     = Strings
}

mapping "aws_opsworks_haproxy_layer" {
  stack_id                    = String
  // stats_password           = String
  name                        = String
  auto_assign_elastic_ips     = Boolean
  auto_assign_public_ips      = Boolean
  custom_instance_profile_arn = String
  custom_security_group_ids   = Strings
  auto_healing                = Boolean
  healthcheck_method          = String
  healthcheck_url             = String
  install_updates_on_boot     = Boolean
  instance_shutdown_timeout   = Integer
  elastic_load_balancer       = String
  drain_elb_on_shutdown       = Boolean
  // stats_enabled            = String
  // stats_url                = String
  // stats_user               = String
  system_packages             = Strings
  use_ebs_optimized_instances = Boolean
  ebs_volume                  = VolumeConfiguration
  custom_json                 = String
  custom_configure_recipes    = Strings
  custom_deploy_recipes       = Strings
  custom_setup_recipes        = Strings
  custom_shutdown_recipes     = Strings
  custom_undeploy_recipes     = Strings
}

mapping "aws_opsworks_instance" {
  instance_type           = String
  stack_id                = String
  layer_ids               = Strings
  state                   = any
  install_updates_on_boot = Boolean
  auto_scaling_type       = AutoScalingType
  availability_zone       = String
  ebs_optimized           = Boolean
  hostname                = String
  architecture            = Architecture
  ami_id                  = String
  os                      = String
  root_device_type        = RootDeviceType
  ssh_key_name            = String
  agent_version           = String
  subnet_id               = String
  tenancy                 = String
  virtualization_type     = String
  root_block_device       = BlockDeviceMapping
  ebs_block_device        = BlockDeviceMapping
  ephemeral_block_device  = BlockDeviceMapping
}

mapping "aws_opsworks_java_app_layer" {
  stack_id                    = String
  name                        = String
  app_server                  = String
  app_server_version          = String
  auto_assign_elastic_ips     = Boolean
  auto_assign_public_ips      = Boolean
  custom_instance_profile_arn = String
  custom_security_group_ids   = Strings
  auto_healing                = Boolean
  install_updates_on_boot     = Boolean
  instance_shutdown_timeout   = Integer
  jvm_type                    = String
  jvm_options                 = String
  jvm_version                 = String
  elastic_load_balancer       = String
  drain_elb_on_shutdown       = Boolean
  system_packages             = Strings
  use_ebs_optimized_instances = Boolean
  ebs_volume                  = VolumeConfiguration
  custom_json                 = String
  custom_configure_recipes    = Strings
  custom_deploy_recipes       = Strings
  custom_setup_recipes        = Strings
  custom_shutdown_recipes     = Strings
  custom_undeploy_recipes     = Strings
}

mapping "aws_opsworks_memcached_layer" {
  stack_id                    = String
  name                        = String
  // allocated_memory         = String
  auto_assign_elastic_ips     = Boolean
  auto_assign_public_ips      = Boolean
  custom_instance_profile_arn = String
  custom_security_group_ids   = Strings
  auto_healing                = Boolean
  install_updates_on_boot     = Boolean
  instance_shutdown_timeout   = Integer
  elastic_load_balancer       = String
  drain_elb_on_shutdown       = Boolean
  system_packages             = Strings
  use_ebs_optimized_instances = Boolean
  ebs_volume                  = VolumeConfiguration
  custom_json                 = String
  custom_configure_recipes    = Strings
  custom_deploy_recipes       = Strings
  custom_setup_recipes        = Strings
  custom_shutdown_recipes     = Strings
  custom_undeploy_recipes     = Strings
}

mapping "aws_opsworks_mysql_layer" {
  stack_id                       = String
  name                           = String
  auto_assign_elastic_ips        = Boolean
  auto_assign_public_ips         = Boolean
  custom_instance_profile_arn    = String
  custom_security_group_ids      = Strings
  auto_healing                   = Boolean
  install_updates_on_boot        = Boolean
  instance_shutdown_timeout      = Integer
  elastic_load_balancer          = String
  drain_elb_on_shutdown          = Boolean
  // root_password                  = String
  // root_password_on_all_instances = String
  system_packages                = Strings
  use_ebs_optimized_instances    = Boolean
  ebs_volume                     = VolumeConfiguration
  custom_json                    = String
  custom_configure_recipes       = Strings
  custom_deploy_recipes          = Strings
  custom_setup_recipes           = Strings
  custom_shutdown_recipes        = Strings
  custom_undeploy_recipes        = Strings
}

mapping "aws_opsworks_nodejs_app_layer" {
  stack_id                    = String
  name                        = String
  auto_assign_elastic_ips     = Boolean
  auto_assign_public_ips      = Boolean
  custom_instance_profile_arn = String
  custom_security_group_ids   = Strings
  auto_healing                = Boolean
  install_updates_on_boot     = Boolean
  instance_shutdown_timeout   = Integer
  elastic_load_balancer       = String
  drain_elb_on_shutdown       = Boolean
  nodejs_version              = String
  system_packages             = Strings
  use_ebs_optimized_instances = Boolean
  ebs_volume                  = VolumeConfiguration
  custom_json                 = String
  custom_configure_recipes    = Strings
  custom_deploy_recipes       = Strings
  custom_setup_recipes        = Strings
  custom_shutdown_recipes     = Strings
  custom_undeploy_recipes     = Strings
}

mapping "aws_opsworks_permission" {
  allow_ssh  = Boolean
  allow_sudo = Boolean
  user_arn   = String
  level      = String
  stack_id   = String
}

mapping "aws_opsworks_php_app_layer" {
  stack_id                    = String
  name                        = String
  auto_assign_elastic_ips     = Boolean
  auto_assign_public_ips      = Boolean
  custom_instance_profile_arn = String
  custom_security_group_ids   = Strings
  auto_healing                = Boolean
  install_updates_on_boot     = Boolean
  instance_shutdown_timeout   = Integer
  elastic_load_balancer       = String
  drain_elb_on_shutdown       = Boolean
  system_packages             = Strings
  use_ebs_optimized_instances = Boolean
  ebs_volume                  = VolumeConfiguration
  custom_json                 = String
  custom_configure_recipes    = Strings
  custom_deploy_recipes       = Strings
  custom_setup_recipes        = Strings
  custom_shutdown_recipes     = Strings
  custom_undeploy_recipes     = Strings
}

mapping "aws_opsworks_rails_app_layer" {
  stack_id                    = String
  name                        = String
  // app_server               = String
  auto_assign_elastic_ips     = Boolean
  auto_assign_public_ips      = Boolean
  // bundler_version          = String
  custom_instance_profile_arn = String
  custom_security_group_ids   = Strings
  auto_healing                = Boolean
  install_updates_on_boot     = Boolean
  instance_shutdown_timeout   = Integer
  elastic_load_balancer       = String
  drain_elb_on_shutdown       = Boolean
  // manage_bundler           = Boolean
  // passenger_version        = String
  // ruby_version             = String
  // rubygems_version         = String
  system_packages             = Strings
  use_ebs_optimized_instances = Boolean
  ebs_volume                  = VolumeConfiguration
  custom_json                 = String
  custom_configure_recipes    = Strings
  custom_deploy_recipes       = Strings
  custom_setup_recipes        = Strings
  custom_shutdown_recipes     = Strings
  custom_undeploy_recipes     = Strings
}

mapping "aws_opsworks_rds_db_instance" {
  stack_id            = String
  rds_db_instance_arn = String
  db_user             = String
  db_password         = String
}

mapping "aws_opsworks_stack" {
  name                          = String
  region                        = String
  service_role_arn              = String
  default_instance_profile_arn  = String
  agent_version                 = String
  berkshelf_version             = String
  color                         = String
  default_availability_zone     = String
  configuration_manager_name    = String
  configuration_manager_version = String
  custom_cookbooks_source       = Source
  custom_json                   = String
  default_os                    = String
  default_root_device_type      = RootDeviceType
  default_ssh_key_name          = String
  default_subnet_id             = String
  hostname_theme                = String
  manage_berkshelf              = Boolean
  tags                          = Tags
  use_custom_cookbooks          = Boolean
  use_opsworks_security_groups  = Boolean
  vpc_id                        = String
}

mapping "aws_opsworks_static_web_layer" {
  stack_id                    = String
  name                        = String
  auto_assign_elastic_ips     = Boolean
  auto_assign_public_ips      = Boolean
  custom_instance_profile_arn = String
  custom_security_group_ids   = Strings
  auto_healing                = Boolean
  install_updates_on_boot     = Boolean
  instance_shutdown_timeout   = Integer
  elastic_load_balancer       = String
  drain_elb_on_shutdown       = Boolean
  system_packages             = Strings
  use_ebs_optimized_instances = Boolean
  ebs_volume                  = VolumeConfiguration
  custom_json                 = String
  custom_configure_recipes    = Strings
  custom_deploy_recipes       = Strings
  custom_setup_recipes        = Strings
  custom_shutdown_recipes     = Strings
  custom_undeploy_recipes     = Strings
}

mapping "aws_opsworks_user_profile" {
  user_arn              = String
  allow_self_management = Boolean
  ssh_username          = String
  ssh_public_key        = String
}
