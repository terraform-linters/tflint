# AWS ElastiCache Cluster Default Parameter Group
Report this issue if you have specified the default parameter group. This issue type is NOTICE.

## Example
```
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "app"
  engine               = "redis"
  engine_version       = "3.2.4"
  maintenance_window   = "sun:00:00-sun:06:00"
  node_type            = "cache.m4.large"
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
        NOTICE:9 "default.redis3.2" is default parameter group. You cannot edit it.

Result: 1 issues  (0 errors , 0 warnings , 1 notices)
```

## Why
In ElastiCache, parameter groups can be used for tuning setting values ​​such as Redis and Memcached. When creating a cluster for the first time, you can only select the default parameter group. However, the default parameter group can not be changed later, and if you want to change it, you need to change it to another parameter group.

## How To Fix
Please create a dedicated parameter group and change it to that.
