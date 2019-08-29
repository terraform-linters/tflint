rule "aws_elasticache_cluster_invalid_parameter_group" {
    resource      = "aws_elasticache_cluster"
    attribute     = "parameter_group_name"
    source_action = "DescribeCacheParameterGroups"
    template      = "\"%s\" is invalid parameter group name."
}

rule "aws_elasticache_cluster_invalid_security_group" {
    resource      = "aws_elasticache_cluster"
    attribute     = "security_group_ids"
    source_action = "DescribeSecurityGroups"
    template      = "\"%s\" is invalid security group."
}

rule "aws_elasticache_cluster_invalid_subnet_group" {
    resource      = "aws_elasticache_cluster"
    attribute     = "subnet_group_name"
    source_action = "DescribeCacheSubnetGroups"
    template      = "\"%s\" is invalid subnet group name."
}
