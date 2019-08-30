rule "aws_launch_configuration_invalid_iam_profile" {
    resource      = "aws_launch_configuration"
    attribute     = "iam_instance_profile"
    source_action = "ListInstanceProfiles"
    template      = "\"%s\" is invalid IAM profile name."
}
