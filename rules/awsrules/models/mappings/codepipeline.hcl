import = "aws-sdk-go/models/apis/codepipeline/2015-07-09/api-2.json"

mapping "aws_codepipeline" {
  name     = PipelineName
  role_arn = RoleArn
}

mapping "aws_codepipeline_webhook" {
  name            = WebhookName
  authentication  = WebhookAuthenticationType
  target_action   = ActionName
  target_pipeline = PipelineName
}

test "aws_codepipeline" "name" {
  ok = "tf-test-pipeline"
  ng = "test/pipeline"
}

test "aws_codepipeline" "role_arn" {
  ok = "arn:aws:iam::123456789012:role/s3access"
  ng = "arn:aws:iam::123456789012:instance-profile/s3access-profile"
}

test "aws_codepipeline_webhook" "name" {
  ok = "test-webhook-github-bar"
  ng = "webhook-github-bar/testing"
}

test "aws_codepipeline_webhook" "authentication" {
  ok = "GITHUB_HMAC"
  ng = "GITLAB_HMAC"
}

test "aws_codepipeline_webhook" "target_action" {
  ok = "Source"
  ng = "Source/Example"
}
