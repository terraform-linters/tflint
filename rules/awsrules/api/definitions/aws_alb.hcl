rule "aws_alb_invalid_security_group" {
    resource      = "aws_alb"
    attribute     = "security_groups"
    source_action = "DescribeSecurityGroups"
    template      = "\"%s\" is invalid security group."
}

rule "aws_alb_invalid_subnet" {
    resource      = "aws_alb"
    attribute     = "subnets"
    source_action = "DescribeSubnets"
    template      = "\"%s\" is invalid subnet ID."
}
