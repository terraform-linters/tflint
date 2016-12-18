# AWS ELB Invalid Instance
Report this issue if you have specified the invalid instance ID. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_elb" "balancer" {
  name = "foobar-terraform-elb"
  subnets = ["subnet-1234abcd"]
  security_groups = ["sg-12345678"]
  instances = ["i-1234abcd"] # This instance ID does not exists

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
        ERROR:5 "i-1234abcd" is invalid instance.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid instance ID is specified, an error will occur at `terraform apply`.

## How to fix
Check your instances and select a valid instance ID again.
