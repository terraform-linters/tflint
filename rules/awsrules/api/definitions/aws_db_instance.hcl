rule "aws_db_instance_invalid_db_subnet_group" {
    resource      = "aws_db_instance"
    attribute     = "db_subnet_group_name"
    source_action = "DescribeDBSubnetGroups"
    template      = "\"%s\" is invalid DB subnet group name."
}

rule "aws_db_instance_invalid_option_group" {
    resource      = "aws_db_instance"
    attribute     = "option_group_name"
    source_action = "DescribeOptionGroups"
    template      = "\"%s\" is invalid option group name."
}

rule "aws_db_instance_invalid_parameter_group" {
    resource      = "aws_db_instance"
    attribute     = "parameter_group_name"
    source_action = "DescribeDBParameterGroups"
    template      = "\"%s\" is invalid parameter group name."
}

rule "aws_db_instance_invalid_vpc_security_group" {
    resource      = "aws_db_instance"
    attribute     = "vpc_security_group_ids"
    source_action = "DescribeSecurityGroups"
    template      = "\"%s\" is invalid security group."
}
