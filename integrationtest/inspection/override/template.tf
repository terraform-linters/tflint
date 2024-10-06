resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t2.micro" // Override by `template_override.tf`
}

terraform {
  backend "s3" {}
}

terraform {
  required_providers {
    aws     = "1" // Override by `template_override.tf`
    google  = "1" // Override by `version_override.tf`
    azurerm = "1"
  }
}
