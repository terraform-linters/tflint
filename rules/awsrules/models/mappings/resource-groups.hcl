import = "aws-sdk-go/models/apis/resource-groups/2017-11-27/api-2.json"

mapping "aws_resourcegroups_group" {
  name           = GroupName
  description    = any //GroupDescription
  resource_query = ResourceQuery
}
