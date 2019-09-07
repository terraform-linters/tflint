import = "aws-sdk-go/models/apis/codebuild/2016-10-06/api-2.json"

mapping "aws_codebuild_project" {
  build_timeout = TimeOut
  description   = ProjectDescription
}

mapping "aws_codebuild_source_credential" {
  auth_type   = AuthType
  server_type = ServerType
  token       = SensitiveNonEmptyString
  user_name   = NonEmptyString
}
