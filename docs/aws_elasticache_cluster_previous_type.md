# AWS ElastiCache Cluster Instance Previous Type
Report this issue if you have specified the previous node type. This issue type is WARNING.

## Example
```
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "app"
  engine               = "redis"
  engine_version       = "3.2.4"
  maintenance_window   = "sun:00:00-sun:06:00"
  node_type            = "cache.t1.micro" // previous type!
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
        WARNING:6 "cache.t1.micro" is previous generation node type.

Result: 1 issues  (0 errors , 1 warnings , 0 notices)
```

## Why
There are two types of node types, the current generation and the previous generation. The current generation is superior to the previous generation in terms of performance and fee. AWS also officially states that unless there is a special reason, you should use the node type of the current generation.

## How To Fix
Follow the [upgrade paths](https://aws.amazon.com/elasticache/previous-generation/) and confirm that the node type of the current generation can be used, then select again.
