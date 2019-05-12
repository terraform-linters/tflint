import = "aws-sdk-go/models/apis/ram/2018-01-04/api-2.json"

mapping "aws_ram_principal_association" {
  principal          = String
  resource_share_arn = String
}

mapping "aws_ram_resource_association" {
  resource_arn       = String
  resource_share_arn = String
}

mapping "aws_ram_resource_share" {
  name                      = String
  allow_external_principals = Boolean
  tags                      = TagList
}
