# aws_elasticache_cluster_previous_type

Disallow using previous node types.

## Example

```hcl
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "app"
  engine               = "redis"
  engine_version       = "3.2.4"
  maintenance_window   = "sun:00:00-sun:06:00"
  node_type            = "cache.t1.micro" // previous node type!
  num_cache_nodes      = 1
  port                 = 6379
  parameter_group_name = "default.redis3.2"
  subnet_group_name    = "app-subnet-group"
  security_group_ids   = ["${aws_security_group.redis.id}"]
}
```

```console
$ tflint
1 issue(s) found:

Warning: "cache.t1.micro" is previous generation node type. (aws_elasticache_cluster_previous_type)

  on template.tf line 6:
   6:   node_type            = "cache.t1.micro" // previous node type!

Reference: https://github.com/wata727/tflint/blob/v0.11.0/docs/rules/aws_elasticache_cluster_previous_type.md

```

## Why

Previous node types are inferior to current generation in terms of performance and fee. Unless there is a special reason, you should avoid to use these ones.

## How To Fix

Select a current generation node type according to the [upgrade paths](https://aws.amazon.com/elasticache/previous-generation/).
