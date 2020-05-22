# terraform_blacklisted_resources

Disallow `resource` declarations of certain types. Each type has a custom message will be displayed.

## Example

```hcl
rule "terraform_blacklisted_resources" {
  enabled = true
  types   = {
    aws_iam_policy_attachment = "Consider aws_iam_role_policy_attachment, aws_iam_user_policy_attachment, or aws_iam_group_policy_attachment instead."
  }
}
```

```hcl
resource "random_id" "server" {
  byte_length = 8
}

resource "aws_iam_policy_attachment" "test-attach" {
  name       = "test-attachment"
  users      = [aws_iam_user.user.name]
  roles      = [aws_iam_role.role.name]
  groups     = [aws_iam_group.group.name]
  policy_arn = aws_iam_policy.policy.arn
}
```

```
$ tflint
1 issue(s) found:

Warning: `aws_iam_policy_attachment` resource type is blacklisted

Consider aws_iam_role_policy_attachment, aws_iam_user_policy_attachment, or aws_iam_group_policy_attachment instead. (terraform_blacklisted_resources)

  on modules.tf line 1:
   1: resource "aws_iam_policy_attachment" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.16.0/docs/rules/terraform_blacklisted_resources.md
```

## Why
Organizations may want to disallow some resource types from being used in their organization and provide feedback to 
alternative resources. An example is the `aws_iam_policy_attachment` and preferring to use
`aws_iam_role_policy_attachment`, `aws_iam_user_policy_attachment`, or `aws_iam_group_policy_attachment` instead.

## How To Fix
Follow the guidance provided by the message associated with the blacklisted resource type.
