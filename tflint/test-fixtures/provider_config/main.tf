variable "secret_key" {}

variable "region" {
    default = "us-east-1"
}

variable "role_arn" {
    default = "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
}

provider "aws" {
    access_key = "AWS_ACCESS_KEY"
    secret_key = var.secret_key
    region     = var.region
    profile    = null
    shared_credentials_file = null_resource.foo.bar

    assume_role {
      role_arn     = var.role_arn
      session_name = "SESSION_NAME"
      external_id  = "EXTERNAL_ID"
      policy       = "POLICY_NAME"
    }
}

resource "null_resource" "foo" {}
