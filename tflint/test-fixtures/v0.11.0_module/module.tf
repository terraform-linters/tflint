module "ecs_on_spotfleet" {
  source = "github.com/wata727/tf_aws_ecs_on_spotfleet.git?ref=master"
}

module "consul" {
   source = "hashicorp/consul/aws"
   version = ">= 0.0.3, <= 0.0.5"
}
