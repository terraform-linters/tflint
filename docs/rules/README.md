# Rules

Rules related to AWS provider and Terraform are available. These rules are enabled by default.

## AWS Rules

These rules relate to AWS provider.

### Possible Errors

These rules warn of possible errors that can occur at `terraform apply`. Rules marked with `Deep` are only used when enabling deep checking:

|Rule|Deep|
| --- | --- |
|aws_alb_invalid_security_group|✔|
|aws_alb_invalid_subnet|✔|
|aws_db_instance_invalid_db_subnet_group|✔|
|aws_db_instance_invalid_option_group|✔|
|aws_db_instance_invalid_parameter_group|✔|
|aws_db_instance_invalid_type||
|aws_db_instance_invalid_vpc_security_group|✔|
|aws_elasticache_cluster_invalid_parameter_group|✔|
|aws_elasticache_cluster_invalid_security_group|✔|
|aws_elasticache_cluster_invalid_subnet_group|✔|
|aws_elasticache_cluster_invalid_type||
|aws_elb_invalid_instance|✔|
|aws_elb_invalid_security_group|✔|
|aws_elb_invalid_subnet|✔|
|aws_instance_invalid_ami|✔|
|aws_instance_invalid_iam_profile|✔|
|aws_instance_invalid_key_name|✔|
|aws_instance_invalid_subnet|✔|
|aws_instance_invalid_vpc_security_group|✔|
|aws_launch_configuration_invalid_iam_profile|✔|
|aws_launch_configuration_invalid_image_id|✔|
|aws_route_invalid_egress_only_gateway|✔|
|aws_route_invalid_gateway|✔|
|aws_route_invalid_instance|✔|
|aws_route_invalid_nat_gateway|✔|
|aws_route_invalid_network_interface|✔|
|aws_route_invalid_route_table|✔|
|aws_route_invalid_vpc_peering_connection|✔|
|[aws_s3_bucket_name](aws_s3_bucket_name.md)||
|[aws_route_not_specified_target](aws_route_not_specified_target.md)||
|[aws_route_specified_multiple_targets](aws_route_specified_multiple_targets.md)||

#### SDK-based Validations

700+ rules based on the aws-sdk validations are also available. See [full list](../../rules/awsrules/models/).

### Best Practices

These rules suggest to better ways.

- [aws_instance_previous_type](aws_instance_previous_type.md)
- [aws_db_instance_previous_type](aws_db_instance_previous_type.md)
- [aws_db_instance_default_parameter_group](aws_db_instance_default_parameter_group.md)
- [aws_elasticache_cluster_previous_type](aws_elasticache_cluster_previous_type.md)
- [aws_elasticache_cluster_default_parameter_group](aws_elasticache_cluster_default_parameter_group.md)

## Terraform Rules

These rules relate to Terraform itself, not providers.

### Best Practices

These rules suggest to better ways.

|Rule|Enabled by default|
| --- | --- |
|[terraform_deprecated_interpolation](terraform_deprecated_interpolation.md)|✔|
|[terraform_deprecated_index](terraform_deprecated_index.md)||
|[terraform_unused_declarations](terraform_unused_declarations.md)||
|[terraform_comment_syntax](terraform_comment_syntax.md)||
|[terraform_documented_outputs](terraform_documented_outputs.md)||
|[terraform_documented_variables](terraform_documented_variables.md)||
|[terraform_typed_variables](terraform_typed_variables.md)||
|[terraform_module_pinned_source](terraform_module_pinned_source.md)|✔|
|[terraform_naming_convention](terraform_naming_convention.md)||
|[terraform_required_version](terraform_required_version.md)||
|[terraform_required_providers](terraform_required_providers.md)||
|[terraform_standard_module_structure](terraform_standard_module_structure.md)||
|[terraform_workspace_remote](terraform_workspace_remote.md)|✔|
