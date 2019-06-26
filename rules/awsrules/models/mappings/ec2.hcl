import = "aws-sdk-go/models/apis/ec2/2016-11-15/api-2.json"

mapping "aws_ami" {
  name                   = String
  description            = String
  ena_support            = Boolean
  root_device_name       = String
  virtualization_type    = String
  architecture           = ArchitectureValues
  ebs_block_device       = BlockDeviceMappingRequestList
  ephemeral_block_device = BlockDeviceMappingRequestList
  tags                   = TagList
  image_location         = String
  kernel_id              = String
  ramdisk_id             = String
  sriov_net_support      = String
}

mapping "aws_ami_copy" {
  name              = String
  source_ami_id     = String
  source_ami_region = String
  encrypted         = Boolean
  kms_key_id        = String
  tags              = TagList
}

mapping "aws_ami_from_instance" {
  name                    = String
  source_instance_id      = String
  snapshot_without_reboot = Boolean
  tags                    = TagList
}

mapping "aws_ami_launch_permission" {
  image_id   = String
  account_id = String
}

mapping "aws_ebs_snapshot" {
  volume_id   = String
  description = String
  tags        = TagList
}

mapping "aws_ebs_snapshot_copy" {
  description        = String
  encrypted          = Boolean
  kms_key_id         = String
  source_snapshot_id = String
  source_region      = String
  tags               = TagList
}

mapping "aws_ebs_volume" {
  availability_zone = String
  encrypted         = Boolean
  iops              = Integer
  size              = Integer
  snapshot_id       = String
  type              = VolumeType
  kms_key_id        = String
  tags              = TagSpecificationList
}

mapping "aws_ec2_capacity_reservation" {
  availability_zone       = String
  ebs_optimized           = Boolean
  end_date                = DateTime
  end_date_type           = EndDateType
  ephemeral_storage       = Boolean
  instance_count          = Integer
  instance_match_criteria = InstanceMatchCriteria
  instance_platform       = CapacityReservationInstancePlatform
  instance_type           = String
  tags                    = TagSpecificationList
  tenancy                 = CapacityReservationTenancy
}

mapping "aws_ec2_client_vpn_endpoint" {
  description            = String
  client_cidr_block      = String
  dns_servers            = ValueStringList
  server_certificate_arn = String
  transport_protocol     = TransportProtocol
  authentication_options = ClientVpnAuthenticationRequestList
  connection_log_options = ConnectionLogOptions
  tags                   = TagSpecificationList
}

mapping "aws_ec2_client_vpn_network_association" {
  client_vpn_endpoint_id = String
  subnet_id              = String
}

mapping "aws_ec2_fleet" {
  launch_template_config              = FleetLaunchTemplateConfigListRequest
  target_capacity_specification       = TargetCapacitySpecificationRequest
  excess_capacity_termination_policy  = FleetExcessCapacityTerminationPolicy
  on_demand_options                   = OnDemandOptionsRequest
  replace_unhealthy_instances         = Boolean
  spot_options                        = SpotOptionsRequest
  tags                                = TagSpecificationList
  terminate_instances                 = Boolean
  terminate_instances_with_expiration = Boolean
  type                                = FleetType
}

mapping "aws_ec2_transit_gateway" {
  amazon_side_asn                 = Long
  auto_accept_shared_attachments  = AutoAcceptSharedAttachmentsValue
  default_route_table_association = DefaultRouteTableAssociationValue
  default_route_table_propagation = DefaultRouteTablePropagationValue
  description                     = String
  dns_support                     = DnsSupportValue
  tags                            = TagSpecificationList
}

mapping "aws_ec2_transit_gateway_route" {
  destination_cidr_block         = String
  transit_gateway_attachment_id  = String
  transit_gateway_route_table_id = String
}

mapping "aws_ec2_transit_gateway_route_table" {
  transit_gateway_id = String
  tags               = TagSpecificationList
}

mapping "aws_ec2_transit_gateway_route_table_association" {
  transit_gateway_attachment_id  = String
  transit_gateway_route_table_id = String
}

mapping "aws_ec2_transit_gateway_route_table_propagation" {
  transit_gateway_attachment_id  = String
  transit_gateway_route_table_id = String
}

mapping "aws_ec2_transit_gateway_vpc_attachment" {
  subnet_ids                                      = ValueStringList
  transit_gateway_id                              = String
  vpc_id                                          = String
  dns_support                                     = DnsSupportValue
  ipv6_support                                    = Ipv6SupportValue
  tags                                            = TagSpecificationList
  transit_gateway_default_route_table_association = Boolean
  transit_gateway_default_route_table_propagation = Boolean
}

mapping "aws_ec2_transit_gateway_vpc_attachment_accepter" {
  transit_gateway_attachment_id                   = String
  transit_gateway_default_route_table_association = Boolean
  transit_gateway_default_route_table_propagation = Boolean
  tags                                            = TagSpecificationList
}

mapping "aws_eip" {
  vpc                       = Boolean
  instance                  = String
  network_interface         = String
  associate_with_private_ip = String
  tags                      = TagSpecificationList
  public_ipv4_pool          = String
}

mapping "aws_eip_association" {
  allocation_id        = String
  allow_reassociation  = Boolean
  instance_id          = String
  network_interface_id = String
  private_ip_address   = String
  public_ip            = String
}

mapping "aws_instance" {
  ami                                  = String
  availability_zone                    = String
  placement_group                      = Placement
  tenancy                              = Tenancy
  host_id                              = String
  cpu_core_count                       = Integer
  cpu_threads_per_core                 = Integer
  ebs_optimized                        = Boolean
  disable_api_termination              = Boolean
  instance_initiated_shutdown_behavior = ShutdownBehavior
  instance_type                        = InstanceType
  key_name                             = String
  get_password_data                    = Boolean
  monitoring                           = RunInstancesMonitoringEnabled
  security_groups                      = SecurityGroupStringList
  vpc_security_group_ids               = SecurityGroupIdStringList
  subnet_id                            = String
  associate_public_ip_address          = Boolean
  private_ip                           = String
  source_dest_check                    = Boolean
  user_data                            = String
  user_data_base64                     = String
  iam_instance_profile                 = IamInstanceProfileSpecification
  ipv6_address_count                   = Integer
  ipv6_addresses                       = InstanceIpv6AddressList
  tags                                 = TagSpecificationList
  volume_tags                          = TagSpecificationList
  root_block_device                    = BlockDeviceMappingRequestList
  ebs_block_device                     = BlockDeviceMappingRequestList
  ephemeral_block_device               = BlockDeviceMappingRequestList
  network_interface                    = InstanceNetworkInterfaceSpecificationList
  credit_specification                 = CreditSpecificationRequest
}

mapping "aws_key_pair" {
  key_name        = String
  key_name_prefix = String
  public_key      = Blob
}

mapping "aws_launch_template" {
  name                                 = LaunchTemplateName
  name_prefix                          = String
  description                          = VersionDescription
  block_device_mappings                = LaunchTemplateBlockDeviceMappingRequestList
  capacity_reservation_specification   = LaunchTemplateCapacityReservationSpecificationRequest
  credit_specification                 = CreditSpecificationRequest
  disable_api_termination              = Boolean
  ebs_optimized                        = Boolean
  elastic_gpu_specifications           = ElasticGpuSpecificationList
  elastic_inference_accelerator        = LaunchTemplateElasticInferenceAcceleratorList
  iam_instance_profile                 = LaunchTemplateIamInstanceProfileSpecificationRequest
  image_id                             = String
  instance_initiated_shutdown_behavior = ShutdownBehavior
  instance_market_options              = LaunchTemplateInstanceMarketOptionsRequest
  instance_type                        = InstanceType
  kernel_id                            = String
  key_name                             = String
  license_specification                = LaunchTemplateLicenseSpecificationListRequest
  monitoring                           = LaunchTemplatesMonitoringRequest
  network_interfaces                   = LaunchTemplateInstanceNetworkInterfaceSpecificationRequestList
  placement                            = LaunchTemplatePlacementRequest
  ram_disk_id                          = String
  security_group_names                 = SecurityGroupStringList
  vpc_security_group_ids               = SecurityGroupIdStringList
  tag_specifications                   = LaunchTemplateTagSpecificationRequestList
  tags                                 = TagSpecificationList
  user_data                            = String
}

mapping "aws_placement_group" {
  name     = String
  strategy = PlacementStrategy
}

mapping "aws_snapshot_create_volume_permission" {
  snapshot_id = String
  account_id  = String
}

mapping "aws_spot_datafeed_subscription" {
  bucket = String
  prefix = String
}

mapping "aws_spot_fleet_request" {
  iam_fleet_role                      = String
  replace_unhealthy_instances         = Boolean
  launch_specification                = RequestSpotLaunchSpecification
  spot_price                          = SpotPrice
  wait_for_fulfillment                = Boolean
  target_capacity                     = Integer
  allocation_strategy                 = AllocationStrategy
  instance_pools_to_use_count         = Integer
  excess_capacity_termination_policy  = FleetExcessCapacityTerminationPolicy
  terminate_instances_with_expiration = Boolean
  instance_interruption_behaviour     = InstanceInterruptionBehavior
  fleet_type                          = FleetType
  valid_until                         = DateTime
  valid_from                          = DateTime
}

mapping "aws_spot_instance_request" {
  spot_price                      = SpotPrice
  wait_for_fulfillment            = Boolean
  spot_type                       = String
  launch_group                    = String
  block_duration_minutes          = Integer
  instance_interruption_behaviour = InstanceInterruptionBehavior
  valid_until                     = DateTime
  valid_from                      = DateTime
  tags                            = TagSpecificationList
}

mapping "aws_volume_attachment" {
  device_name  = String
  instance_id  = String
  volume_id    = String
  force_detach = Boolean
  skip_destroy = Boolean
}

test "aws_ami" "architecture" {
  ok = "x86_64"
  ng = "x86"
}

test "aws_ebs_volume" "type" {
  ok = "gp2"
  ng = "gp3"
}

test "aws_ec2_capacity_reservation" "end_date_type" {
  ok = "unlimited"
  ng = "unlimit"
}

test "aws_ec2_capacity_reservation" "instance_match_criteria" {
  ok = "open"
  ng = "close"
}

test "aws_ec2_capacity_reservation" "instance_platform" {
  ok = "Linux/UNIX"
  ng = "Linux/GNU"
}

test "aws_ec2_capacity_reservation" "tenancy" {
  ok = "default"
  ng = "reserved"
}

test "aws_ec2_client_vpn_endpoint" "transport_protocol" {
  ok = "udp"
  ng = "http"
}

test "aws_ec2_fleet" "excess_capacity_termination_policy" {
  ok = "termination"
  ng = "remain"
}

test "aws_ec2_fleet" "type" {
  ok = "maintain"
  ng = "remain"
}

test "aws_ec2_transit_gateway" "auto_accept_shared_attachments" {
  ok = "enable"
  ng = "true"
}

test "aws_ec2_transit_gateway" "default_route_table_association" {
  ok = "disable"
  ng = "false"
}

test "aws_ec2_transit_gateway" "default_route_table_propagation" {
  ok = "disable"
  ng = "disabled"
}

test "aws_ec2_transit_gateway" "dns_support" {
  ok = "enable"
  ng = "enabled"
}

test "aws_ec2_transit_gateway_vpc_attachment" "ipv6_support" {
  ok = "enable"
  ng = "on"
}

test "aws_instance" "instance_initiated_shutdown_behavior" {
  ok = "stop"
  ng = "restart"
}

test "aws_instance" "tenancy" {
  ok = "host"
  ng = "server"
}

test "aws_launch_template" "name" {
  ok = "foo"
  ng = "foo[bar]"
}

test "aws_launch_template" "instance_type" {
  ok = "t2.micro"
  ng = "t1.2xlarge"
}

test "aws_placement_group" "strategy" {
  ok = "cluster"
  ng = "instance"
}

test "aws_spot_fleet_request" "allocation_strategy" {
  ok = "lowestPrice"
  ng = "highestPrice"
}

test "aws_spot_fleet_request" "instance_interruption_behaviour" {
  ok = "hibernate"
  ng = "restart"
}
