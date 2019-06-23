import = "aws-sdk-go/models/apis/states/2016-11-23/api-2.json"

mapping "aws_sfn_activity" {
  name = Name
  tags = TagList
}

mapping "aws_sfn_state_machine" {
  name       = Name
  definition = Definition
  role_arn   = Arn
  tags       = TagList
}
