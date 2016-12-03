# AWS Instance Invalid AMI
Report this issue if you have specified the invalid AMI ID. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_instance" "web" {
  ami                  = "ami-1234abcd" # This AMI ID does not exist
  instance_type        = "t2.micro"
  iam_instance_profile = "app-user"

  tags {
    Name = "HelloWorld"
  }
}
```

The following is the execution result of TFLint: 

```
$ tflint --deep
template.tf
        ERROR:2 "ami-1234abcd" is invalid AMI.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid AMI ID is specified, an error will occur at `terraform apply`.

## How to fix
Check your AMIs and select a valid AMI ID again.
