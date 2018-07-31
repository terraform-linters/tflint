# How to Publish AMIs in a New AWS Account

This readme discusses how to migrate the `publish-amis.sh` script to a new AWS account.

To make using this Module as easy as possible, we want to automatically build and publish AMIs based on the 
[/examples/consul-ami/consul.json](https://github.com/hashicorp/terraform-aws-consul/tree/master/examples/consul-ami/consul.json) Packer template upon every release of this repo. 
This way, users can simply git clone this repo and `terraform apply` the [consul-cluster](https://github.com/hashicorp/terraform-aws-consul/tree/master/MAIN.md)
without first having to build their own AMI. Note that the auto-built AMIs are meant mostly for first-time users to 
easily try out a Module. In a production setting, many users will want to validate the contents of their AMI by
manually building it in their own account.

Unfortunately, auto-building AMIs creates a chicken-and-egg problem. How can we run code that automatically finds the
latest AMI until that AMI actually exists? But to build those AMIs, we have to run a build in CircleCI, which also runs
automated tests, which will fail when they cannot find the desired AMI. 

Our solution is that, for the `publish-amis` git branch only, on every commit, we will build and publish AMIs but we will
not run tests. For all other branches, AMIs will only be built upon a new git tag (GitHub release), and tests will be
run on every commit as usual. These settings are configured in the [circle.yml](https://github.com/hashicorp/terraform-aws-consul/tree/master/circle.yml) file.

In addition to the above, don't forget to update the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment 
variables in CircleCI to reflect the new AWS account.

Finally, note that, on a brand new account, many AWS regions are limited to just 5 EC2 Instances in an Auto Scaling Group,
but the automated tests in this repo create up to 10 EC2 Instances. Therefore, automated tests will fail if they run in
a region with too small a limit. To avoid this issue, request an increase in the number of t2-family EC2 Instances 
allowed in every AWS region from AWS support.