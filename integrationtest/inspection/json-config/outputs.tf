# Standard module structure test file
# This file should satisfy the terraform_standard_module_structure rule

output "vpc_id" {
  description = "The ID of the VPC"
  value       = "vpc-12345678"
}

output "instance_id" {
  description = "The ID of the EC2 instance"
  value       = aws_instance.example_instance.id
}