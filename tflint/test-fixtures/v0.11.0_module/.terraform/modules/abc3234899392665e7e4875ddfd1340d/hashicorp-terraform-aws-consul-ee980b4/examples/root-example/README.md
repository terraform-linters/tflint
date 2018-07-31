# Consul Cluster Example

This folder shows an example of Terraform code that uses the [consul-cluster](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/consul-cluster) module to deploy 
a [Consul](https://www.consul.io/) cluster in [AWS](https://aws.amazon.com/). The cluster consists of two Auto Scaling
Groups (ASGs): one with a small number of Consul server nodes, which are responsible for being part of the [consensus 
quorum](https://www.consul.io/docs/internals/consensus.html), and one with a larger number of client nodes, which 
would typically run alongside your apps:

![Consul architecture](https://github.com/hashicorp/terraform-aws-consul/blob/master/_docs/architecture.png?raw=true)

You will need to create an [Amazon Machine Image (AMI)](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) 
that has Consul installed, which you can do using the [consul-ami example](https://github.com/hashicorp/terraform-aws-consul/tree/master/examples/consul-ami)). Note that to keep 
this example simple, both the server ASG and client ASG are running the exact same AMI. In real-world usage, you'd 
probably have multiple client ASGs, and each of those ASGs would run a different AMI that has the Consul agent 
installed alongside your apps.

For more info on how the Consul cluster works, check out the [consul-cluster](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/consul-cluster) documentation.



## Quick start

To deploy a Consul Cluster:

1. `git clone` this repo to your computer.
1. Build a Consul AMI. See the [consul-ami example](https://github.com/hashicorp/terraform-aws-consul/tree/master/examples/consul-ami) documentation for instructions. Make sure to
   note down the ID of the AMI.
1. Install [Terraform](https://www.terraform.io/).
1. Open `vars.tf`, set the environment variables specified at the top of the file, and fill in any other variables that
   don't have a default, including putting your AMI ID into the `ami_id` variable.
1. Run `terraform get`.
1. Run `terraform plan`.
1. If the plan looks good, run `terraform apply`.
1. Run the [consul-examples-helper.sh script](https://github.com/hashicorp/terraform-aws-consul/tree/master/examples/consul-examples-helper/consul-examples-helper.sh) to 
   print out the IP addresses of the Consul servers and some example commands you can run to interact with the cluster:
   `../consul-examples-helper/consul-examples-helper.sh`.

