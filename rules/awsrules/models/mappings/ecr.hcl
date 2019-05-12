import = "aws-sdk-go/models/apis/ecr/2015-09-21/api-2.json"

mapping "aws_ecr_lifecycle_policy" {
  repository = RepositoryName
  policy     = LifecyclePolicyText
}

mapping "aws_ecr_repository" {
  name = RepositoryName
  tags = TagList
}

mapping "aws_ecr_repository_policy" {
  repository = RepositoryName
  policy     = RepositoryPolicyText
}

test "aws_ecr_lifecycle_policy" "repository" {
  ok = "example"
  ng = "example@com"
}
