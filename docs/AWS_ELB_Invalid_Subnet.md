# AWS ELB Invalid Subnet
Report this issue if you have specified the invalid subnet ID. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_elb" "balancer" {
  name = "foobar-terraform-elb"
  subnets = ["subnet-1234abcd"] # This subnet ID does not exists
  security_groups = [
    "sg-12345678"
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
        ERROR:3 "subnet-1234abcd" is invalid subnet ID.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid subnet ID is specified, an error will occur at `terraform apply`.

## How to fix
Check your subnets and select a valid subnet ID again.
