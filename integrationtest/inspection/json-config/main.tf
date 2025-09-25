terraform {
  required_version = ">= 1.0"
}


# This output lacks documentation (but rule is disabled in config)
output "undocumented_output" {
  value = "test"
}

# This output has documentation
output "documented_output" {
  description = "A properly documented output"
  value       = var.proper_variable
}

# Module with pinned source - should pass with flexible style
module "pinned_module" {
  source = "terraform-aws-modules/vpc/aws"
  version = "3.14.0"
}

# Module without pinned source - should fail
module "unpinned_module" {
  source = "./modules/local"
}

# Module that should be ignored
module "ignored_module" {
  source = "./ignore"
}

# Resource with proper naming
resource "aws_instance" "example_instance" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}

# Resource with improper naming (mixed case)
resource "aws_s3_bucket" "TestBucket" {
  bucket = "my-test-bucket"
}

# Data source with improper naming
data "aws_ami" "LatestAmi" {
  most_recent = true
  owners      = ["amazon"]
}

# Local value with improper naming
locals {
  MixedCaseLocal = "value"
  proper_local   = "another_value"
}