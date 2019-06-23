import = "aws-sdk-go/models/apis/batch/2016-08-10/api-2.json"

mapping "aws_batch_compute_environment" {
  state = CEState
  type  = CEType
}

mapping "aws_batch_job_definition" {
  type = JobDefinitionType
}

mapping "aws_batch_job_queue" {
  state = JQState
}

test "aws_batch_compute_environment" "state" {
  ok = "ENABLED"
  ng = "ON"
}

test "aws_batch_compute_environment" "type" {
  ok = "MANAGED"
  ng = "CONTROLLED"
}

test "aws_batch_job_definition" "type" {
  ok = "container"
  ng = "docker"
}

test "aws_batch_job_queue" "state" {
  ok = "ENABLED"
  ng = "ON"
}
