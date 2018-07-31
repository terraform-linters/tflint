# Amazon ECS on Spot Fleet Terraform module

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

A terraform module for create ECS on Spot Fleet. This is a demo repository.
The outline is as following:

* Bid on Spot Fleet and launch instances that spans two AZs.
* Started instances constitute an ECS cluster.
* Invoked containers support dynamic port mapping by ALB.

## Quick Start

By using the bundled ruby script, you can try ECS on Spot Fleet fastest.

```
$ git clone https://github.com/wata727/tf_aws_ecs_on_spotfleet.git
$ cd tf_aws_ecs_on_spotfleet/cli
$ bundle install
$ ruby wizard.rb generate
      create template.tf
$ terraform init
$ terraform apply
```

This script generates Terraform template. By default, it requests the cheapest spot price with the two subnets in default VPC on `us-east-1`. Also, if you do not have a key pair in us-east-1, it will automatically generate `demo-app.pem`. Since AWS credentials are required for this operation, please use environment variables or shared credentials.

If you want to delete this cluster, please run the following:

```
$ terraform destroy
```

## Module Input Variables

**Required**

* `vpc` - VPC id for ECS cluster
* `subnets` - List of subnet ids for ECS cluster, please choose 2 subnets
* `key_name` - Name of key pair for SSH login to ECS cluster instances

**Optional**

* `ami` - ECS cluster instance AMI id, default is Amazon ECS-optimized AMI in `us-east-1`
* `app_name` - Your application name, default is `demo-app`
* `image` - Your docker image name, default it ECS PHP Simple App
* `container_port` - Port number exposed by container, default is 80
* `service_count` - Number of containers, default is 3
* `cpu_unit` - Number of cpu_units for container, default is 128
* `memory` - Number of memory for container, default is 128
* `spot_prices` - Bid amount to spot fleet, please choose 2 prices, default is `$0.03`
* `strategy` - Instance placement strategy name, default is `diversified`
* `instance_count` - Number of instances, default is 3
* `instance_type` - Instance type launched by Spot Fleet. default is `m3.medium`
* `volume_size` - Root volume size, default is 16
* `https` - Whether the load balancer should listen to https requests, default is `false`
* `app_certificate_arn` - The ARN of the ssl certificate, default is empty
* `app_ssl_policy` - The ssl policy, default is `ELBSecurityPolicy-2015-05`
* `valid_until` - limit of Spot Fleet request, default is `2020-12-15T00:00:00Z`

## Usage

Like other modules, you can easily start ECS cluster by adding this module to your template with required parameters.

```hcl
provider "aws" {
  region = "us-east-1"
}

module "ecs_on_spotfleet" {
  source = "github.com/wata727/tf_aws_ecs_on_spotfleet"

  vpc         = "vpc-12345"
  subnets     = ["subnet-12345", "subnet-abcde"]
  spot_prices = ["0.03", "0.02"]
  key_name    = "demo-app"
}

output "endpoint" {
  value = "${module.ecs_on_spotfleet.endpoint}"
}
```

## Customize

This module is very simple, please remodel and create your own module.

## Author

[Kazuma Watanabe](https://github.com/wata727)
