# TFLint issues documentation
## Issue Type
Issues are classified into the following three types.

- ERROR
    - There is a issue that an execution error occurs. We strongly recommend that you fix it if this is reported.
- WARNING
    - It is reported if something that is explicitly unrecommended is used. We recommend that you fix it.
- NOTICE
    - It is a issue that is considered desirable to fix.

## Issues

- AWS Instance
    - [aws_instance_invalid_type](aws_instance_invalid_type.md)
    - [aws_instance_previous_type](aws_instance_previous_type.md)
    - [aws_instance_not_specified_iam_profile](aws_instance_not_specified_iam_profile.md)
    - [aws_instance_default_standard_volume](aws_instance_default_standard_volume.md)
- AWS DB Instance
    - [aws_db_instance_invalid_type](aws_db_instance_invalid_type.md)
    - [aws_db_instance_previous_type](aws_db_instance_previous_type.md)
    - [aws_db_instance_default_parameter_group](aws_db_instance_default_parameter_group.md)
- AWS ElastiCache Cluster
    - [aws_elasticache_cluster_invalid_type](aws_elasticache_cluster_invalid_type.md)
    - [aws_elasticache_cluster_default_parameter_group](aws_elasticache_cluster_default_parameter_group.md)

### Invalid Issues
If you have enabled deep check, you can check if nonexistent values ​​are not used. All issues are reported as ERROR.

- AWS Instance
    - aws_instance_invalid_iam_profile
    - aws_instance_invalid_ami
    - aws_instance_invalid_key_name
    - aws_instance_invalid_subnet
    - aws_instance_invalid_vpc_security_group
- AWS ALB
    - aws_alb_invalid_security_group
    - aws_alb_invalid_subnet
- AWS ELB
    - aws_elb_invalid_security_group
    - aws_elb_invalid_subnet
    - aws_elb_invalid_instance
- AWS DB Instance
    - aws_db_instance_invalid_vpc_security_group
    - aws_db_instance_invalid_db_subnet_group
    - aws_db_instance_invalid_parameter_group
    - aws_db_instance_invalid_option_group
- AWS ElastiCache Cluster
    - aws_elasticache_cluster_invalid_parameter_group
    - aws_elasticache_cluster_invalid_subnet_group
    - aws_elasticache_cluster_invalid_security_group
