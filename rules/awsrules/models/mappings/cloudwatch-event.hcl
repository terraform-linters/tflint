import = "aws-sdk-go/models/apis/events/2015-10-07/api-2.json"

mapping "aws_cloudwatch_event_permission" {
  principal    = Principal
  statement_id = StatementId
  action       = Action
}

mapping "aws_cloudwatch_event_rule" {
  name                = RuleName
  schedule_expression = ScheduleExpression
  description         = RuleDescription
  role_arn            = RoleArn
}

mapping "aws_cloudwatch_event_target" {
  rule       = RuleName
  target_id  = TargetId
  arn        = TargetArn
  input      = TargetInput
  input_path = TargetInputPath
  role_arn   = RoleArn
}

test "aws_cloudwatch_event_permission" "principal" {
  ok = "*"
  ng = "-"
}

test "aws_cloudwatch_event_permission" "statement_id" {
  ok = "OrganizationAccess"
  ng = "Organization Access"
}

test "aws_cloudwatch_event_permission" "action" {
  ok = "events:PutEvents"
  ng = "cloudwatchevents:PutEvents"
}

test "aws_cloudwatch_event_rule" "name" {
  ok = "capture-aws-sign-in"
  ng = "capture aws sign in"
}

test "aws_cloudwatch_event_target" "target_id" {
  ok = "run-scheduled-task-every-hour"
  ng = "run scheduled task every hour"
}
