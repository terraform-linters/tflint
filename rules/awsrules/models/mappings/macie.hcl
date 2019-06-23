import = "aws-sdk-go/models/apis/macie/2017-12-19/api-2.json"

mapping "aws_macie_member_account_association" {
  member_account_id = AWSAccountId
}

mapping "aws_macie_s3_bucket_association" {
  bucket_name         = BucketName
  classification_type = ClassificationType
  member_account_id   = AWSAccountId
  prefix              = Prefix
}
