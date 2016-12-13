# AWS Instance Invalid VPC Security Group
Report this issue if you have specified the invalid security group ID in VPC. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_instance" "web" {
  ami                    = "ami-1234abcd"
  instance_type          = "t2.micro"
  iam_instance_profile   = "app-user"
  key_name               = "secret"
  subnet_id              = "subnet-1234abcd"
  vpc_security_group_ids = [
    "sg-1234abcd",
    "sg-12345678", # This security group ID does not exists
  ]

  tags {
    Name = "HelloWorld"
  }
}
```

The following is the execution result of TFLint: 

```
$ tflint --deep
template.tf
        ERROR:9 "sg-12345678" is invalid security group.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid security group is specified, an error will occur at `terraform apply`.

## How to fix
Check your security groups and select a valid security group ID again.
