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
template.tf
        WARNING:3 "t1.micro" is previous generation instance type. (aws_instance_previous_type)

Result: 1 issues  (0 errors , 1 warnings , 0 notices)
```

## Why

Previous generation instance types are inferior to current generation in terms of performance and fee. Unless there is a special reason, you should avoid to use these ones.

## How To Fix

Select a current generation instance type according to the [upgrade paths](https://aws.amazon.com/ec2/previous-generation/).
