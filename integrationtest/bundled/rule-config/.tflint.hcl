plugin "aws" {
  enabled = true
}

rule "aws_s3_bucket_name" {
  enabled = true
  regex = "^[a-z\\-]+$"
}
