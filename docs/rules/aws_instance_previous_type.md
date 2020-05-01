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

Current generation instance types have better performance and lower cost than previous generations. Users should avoid previous generation instance types, especially for new instances.

## How To Fix

Select a current generation instance type according to the [upgrade paths](https://aws.amazon.com/ec2/previous-generation/).
