config {
  terraform_version = "0.9.11"
  deep_check = true

  aws_credentials = {
    access_key = "AWS_ACCESS_KEY"
    secret_key = "AWS_SECRET_KEY"
    region     = "us-east-1"
  }
}
