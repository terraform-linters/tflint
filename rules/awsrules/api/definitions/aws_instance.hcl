rule "aws_instance_invalid_iam_profile" {
    resource      = "aws_instance"
    attribute     = "iam_instance_profile"
    source_action = "ListInstanceProfiles"
    template      = "\"%s\" is invalid IAM profile name."
}

rule "aws_instance_invalid_key_name" {
    resource      = "aws_instance"
    attribute     = "key_name"
    source_action = "DescribeKeyPairs"
    template      = "\"%s\" is invalid key name."
}

rule "aws_instance_invalid_subnet" {
    resource      = "aws_instance"
    attribute     = "subnet_id"
    source_action = "DescribeSubnets"
    template      = "\"%s\" is invalid subnet ID."
}

rule "aws_instance_invalid_vpc_security_group" {
    resource      = "aws_instance"
    attribute     = "vpc_security_group_ids"
    source_action = "DescribeSecurityGroups"
    template      = "\"%s\" is invalid security group."
}
