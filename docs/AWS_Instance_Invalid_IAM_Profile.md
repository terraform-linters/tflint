# AWS Instance Invalid IAM Profile
Report this issue if you have specified the invalid IAM profile. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_instance" "web" {
  ami                  = "ami-b73b63a0"
  instance_type        = "m2.2xlarge"
  iam_instance_profile = "invalid_profile" # This profile does not exist

  tags {
    Name = "HelloWorld"
  }
}
```

The following is the execution result of TFLint: 

```
$ tflint --deep
template.tf
        ERROR:4 "invalid_profile" is invalid IAM profile name.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid instance IAM profile is specified, an error will occur at `terraform apply`.

## How to fix
Check your IAM profile list and select a valid IAM profile again.
