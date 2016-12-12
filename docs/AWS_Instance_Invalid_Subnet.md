# AWS Instance Invalid Subnet
Report this issue if you have specified the invalid subnet ID. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_instance" "web" {
  ami                  = "ami-1234abcd"
  instance_type        = "t2.micro"
  iam_instance_profile = "app-user"
  key_name             = "secret"
  subnet_id            = "subnet-1234abcd" # This subnet ID does not exists

  tags {
    Name = "HelloWorld"
  }
}
```

The following is the execution result of TFLint: 

```
$ tflint --deep
template.tf
        ERROR:6 "subnet-1234abcd" is invalid subnet ID.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid subnet ID is specified, an error will occur at `terraform apply`.

## How to fix
Check your subnets and select a valid subnet ID again.
