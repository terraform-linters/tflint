variable "secret_key" {}

variable "region" {
    default = "us-east-1"
}

provider "aws" {
    access_key = "AWS_ACCESS_KEY"
    secret_key = var.secret_key
    region     = var.region
    profile    = null
    shared_credentials_file = null_resource.foo.bar
}

resource "null_resource" "foo" {}
