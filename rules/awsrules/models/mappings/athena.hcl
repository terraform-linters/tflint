import = "aws-sdk-go/models/apis/athena/2017-05-18/api-2.json"

mapping "aws_athena_database" {
  name = DatabaseString
}

mapping "aws_athena_named_query" {
  name        = NameString
  database    = DatabaseString
  query       = QueryString
  description = DescriptionString
}

mapping "aws_athena_workgroup" {
  name          = WorkGroupName
  configuration = WorkGroupConfiguration
  description   = WorkGroupDescriptionString
  state         = WorkGroupState
  tags          = TagList
}
