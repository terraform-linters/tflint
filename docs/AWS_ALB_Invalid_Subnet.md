# AWS Instance Invalid Subnet
Report this issue if you have specified the invalid subnet ID. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_alb" "balancer" {
  name            = "test-alb-tf"
  internal        = false
  security_groups = ["sg-12345678"]
  subnets         = [
    "subnet-1234abcd", # This subnet does not exists
    "subnet-abcd1234",
  ]

  enable_deletion_protection = true

  access_logs {
    bucket = "${aws_s3_bucket.alb_logs.bucket}"
    prefix = "test-alb"
  }

  tags {
    Environment = "production"
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
