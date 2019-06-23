import = "aws-sdk-go/models/apis/dlm/2018-01-12/api-2.json"

mapping "aws_dlm_lifecycle_policy" {
  description        = PolicyDescription
  execution_role_arn = ExecutionRoleArn
  policy_details     = PolicyDetails
  state              = SettablePolicyStateValues
}

test "aws_dlm_lifecycle_policy" "state" {
  ok = "ENABLED"
  ng = "ERROR"
}
