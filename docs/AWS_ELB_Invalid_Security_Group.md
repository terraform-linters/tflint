# AWS ELB Invalid Security Group
Report this issue if you have specified the invalid security group ID in VPC. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_elb" "balancer" {
  name = "foobar-terraform-elb"
  security_groups = [
    "sg-12345678" # This security group ID does not exists
  ]

  access_logs {
    bucket = "foo"
    bucket_prefix = "bar"
    interval = 60
  }

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
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
