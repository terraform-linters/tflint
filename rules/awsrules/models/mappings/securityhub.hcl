import = "aws-sdk-go/models/apis/securityhub/2018-10-26/api-2.json"

mapping "aws_securityhub_account" {}

mapping "aws_securityhub_product_subscription" {
  product_arn = NonEmptyString
}

mapping "aws_securityhub_standards_subscription" {
  standards_arn = NonEmptyString
}
