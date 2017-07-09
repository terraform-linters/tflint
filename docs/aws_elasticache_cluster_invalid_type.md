# AWS ElastiCache Cluster Invalid Type
Report this issue if you have specified the invalid node type. This issue type is ERROR.

## Example
```
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "app"
  engine               = "redis"
  engine_version       = "3.2.4"
  maintenance_window   = "sun:00:00-sun:06:00"
  node_type            = "m4.large" // invalid type!
  num_cache_nodes      = 1
  port                 = 6379
  parameter_group_name = "default.redis3.2"
  subnet_group_name    = "app-subnet-group"
  security_group_ids   = ["${aws_security_group.redis.id}"]
}
```

The following is the execution result of TFLint:

```
$ tflint
template.tf
        ERROR:6 "m4.large" is invalid node type. (aws_elasticache_cluster_invalid_type)

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid node type is specified, an error will occur at `terraform apply`.

## How To Fix
Check the [node type list](https://aws.amazon.com/elasticache/details/) and select a valid node type again.
