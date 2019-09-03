import = "aws-sdk-go/models/apis/ds/2015-04-16/api-2.json"

mapping "aws_directory_service_directory" {
  name             = DirectoryName
  password         = ConnectPassword
  size             = DirectorySize
  vpc_settings     = DirectoryVpcSettings
  connect_settings = DirectoryConnectSettings
  // alias         = AliasName
  description      = Description
  short_name       = DirectoryShortName
  type             = DirectoryType
  edition          = DirectoryEdition
  tags             = Tags
}

mapping "aws_directory_service_conditional_forwarder" {
  directory_id       = DirectoryId
  dns_ips            = DnsIpAddrs
  remote_domain_name = RemoteDomainName
}

mapping "aws_directory_service_log_subscription" {
  directory_id   = DirectoryId
  log_group_name = LogGroupName
}

test "aws_directory_service_directory" "name" {
  ok = "corp.notexample.com"
  ng = "@example.com"
}

test "aws_directory_service_directory" "size" {
  ok = "Small"
  ng = "Micro"
}

test "aws_directory_service_directory" "short_name" {
  ok = "CORP"
  ng = "CORP:EXAMPLE"
}

test "aws_directory_service_directory" "description" {
  ok = "example"
  ng = "@example"
}

test "aws_directory_service_directory" "type" {
  ok = "SimpleAD"
  ng = "ActiveDirectory"
}

test "aws_directory_service_directory" "edition" {
  ok = "Enterprise"
  ng = "Free"
}

test "aws_directory_service_conditional_forwarder" "directory_id" {
  ok = "d-1234567890"
  ng = "1234567890"
}

test "aws_directory_service_conditional_forwarder" "remote_domain_name" {
  ok = "example.com"
  ng = "example^com"
}
