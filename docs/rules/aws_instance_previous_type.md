# aws_instance_previous_type

Disallow using previous generation instance types.

## Example

```hcl
resource "aws_instance" "web" {
  ami                  = "ami-b73b63a0"
  instance_type        = "t1.micro" # previous instance type!
  iam_instance_profile = "app-service"

  tags {
    Name = "HelloWorld"
  }
}
```

```
$ tflint
1 issue(s) found:

Warning: "t1.micro" is previous generation instance type. (aws_instance_previous_type)

  on template.tf line 3:
   3:   instance_type        = "t1.micro" # previous instance type!

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/aws_instance_previous_type.md
 
```

## Why

Previous generation instance types are inferior to current generation in terms of performance and fee. Unless there is a special reason, you should avoid to use these ones.

## How To Fix

Select a current generation instance type according to the [upgrade paths](https://aws.amazon.com/ec2/previous-generation/).
