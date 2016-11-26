# AWS Instance Invalid Type
Report this issue if you have specified the invalid instance type. This issue type is ERROR.

## Example
```
resource "aws_instance" "web" {
  ami                  = "ami-b73b63a0"
  instance_type        = "t2.2xlarge" # invalid type!
  iam_instance_profile = "app-service"

  tags {
    Name = "HelloWorld"
  }
}
```

The following is the execution result of TFLint: 

```
$ tflint
template.tf
        ERROR:3 "t2.2xlarge" is invalid instance type.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid instance type is specified, an error will occur at `terraform apply`.

## How to fix
Check the [instance type list](https://aws.amazon.com/ec2/instance-types/) and select a valid instance type again.
