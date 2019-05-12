import = "aws-sdk-go/models/apis/codecommit/2015-04-13/api-2.json"

mapping "aws_codecommit_repository" {
  repository_name = RepositoryName
  description     = RepositoryDescription
  default_branch  = BranchName
}

mapping "aws_codecommit_trigger" {
  repository_name = RepositoryName
}

test "aws_codecommit_repository" "repository_name" {
  ok = "MyTestRepository"
  ng = "mytest@repository"
}
