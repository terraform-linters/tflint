variable "redis_invalid" {}
variable "redis_previous" {}

resource "aws_elasticache_cluster" "invalid" {
  cluster_id           = "app"
  engine               = "redis"
  engine_version       = "3.2.4"
  maintenance_window   = "sun:00:00-sun:06:00"
  node_type            = "${var.redis_invalid}" // aws_elasticache_cluster_invalid_type
  num_cache_nodes      = 1
  port                 = 6379
  parameter_group_name = "default.redis3.2"     // aws_elasticache_cluster_default_parameter_group
  subnet_group_name    = "app-subnet-group"
  security_group_ids   = ["sg-12345678"]
}

resource "aws_elasticache_cluster" "previous" {
  cluster_id           = "app"
  engine               = "redis"
  engine_version       = "3.2.4"
  maintenance_window   = "sun:00:00-sun:06:00"
  node_type            = "${var.redis_previous}" // aws_elasticache_cluster_previous_type
  num_cache_nodes      = 1
  port                 = 6379
  parameter_group_name = "application.redis3.2"
  subnet_group_name    = "app-subnet-group"
  security_group_ids   = ["sg-12345678"]
}
