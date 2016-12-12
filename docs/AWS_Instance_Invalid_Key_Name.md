# AWS Instance Invalid Key Name
Report this issue if you have specified the invalid key name. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_instance" "web" {
  ami                  = "ami-1234abcd"
  instance_type        = "t2.micro"
  iam_instance_profile = "app-user"
  key_name             = "secret" # This key name does not exists 

  tags {
    Name = "HelloWorld"
  }
}
```

The following is the execution result of TFLint: 

```
$ tflint --deep
template.tf
        ERROR:5 "secret" is invalid key name.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid key name is specified, an error will occur at `terraform apply`.

## How to fix
Check your key pairs and select a valid key name again.
