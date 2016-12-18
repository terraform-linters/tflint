# AWS ALB Invalid Security Group
Report this issue if you have specified the invalid security group ID in VPC. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_alb" "balancer" {
  name            = "test-alb-tf"
  internal        = false
  security_groups = ["sg-12345678"] # This security group does not exists
  subnets         = ["${aws_subnet.public.*.id}"]

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
        ERROR:4 "sg-12345678" is invalid security group.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid security group is specified, an error will occur at `terraform apply`.

## How to fix
Check your security groups and select a valid security group ID again.
