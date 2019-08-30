rule "aws_elb_invalid_instance" {
    resource      = "aws_elb"
    attribute     = "instances"
    source_action = "DescribeInstances"
    template      = "\"%s\" is invalid instance."
}

rule "aws_elb_invalid_security_group" {
    resource      = "aws_elb"
    attribute     = "security_groups"
    source_action = "DescribeSecurityGroups"
    template      = "\"%s\" is invalid security group."
}

rule "aws_elb_invalid_subnet" {
    resource      = "aws_elb"
    attribute     = "subnets"
    source_action = "DescribeSubnets"
    template      = "\"%s\" is invalid subnet ID."
}
