# Consul AMI

This folder shows an example of how to use the [install-consul](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/install-consul) and 
[install-dnsmasq](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/install-dnsmasq) modules with [Packer](https://www.packer.io/) to create [Amazon Machine 
Images (AMIs)](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) that have Consul and Dnsmasq installed on 
top of:
 
1. Ubuntu 16.04
1. Amazon Linux

These AMIs will have [Consul](https://www.consul.io/) installed and configured to automatically join a cluster during 
boot-up. They also have [Dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) installed and configured to use 
Consul for DNS lookups of the `.consul` domain (e.g. `foo.service.consul`) (see [registering 
services](https://www.consul.io/intro/getting-started/services.html) for instructions on how to register your services
in Consul). To see how to deploy this AMI, check out the [consul-cluster example](https://github.com/hashicorp/terraform-aws-consul/tree/master/MAIN.md). 

For more info on Consul installation and configuration, check out the 
[install-consul](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/install-consul) and [install-dnsmasq](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/install-dnsmasq) documentation.



## Quick start

To build the Consul AMI:

1. `git clone` this repo to your computer.
1. Install [Packer](https://www.packer.io/).
1. Configure your AWS credentials using one of the [options supported by the AWS 
   SDK](http://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html). Usually, the easiest option is to
   set the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.
1. Update the `variables` section of the `consul.json` Packer template to configure the AWS region, Consul version, and 
   Dnsmasq version you wish to use.
1. Run `packer build consul.json`.

When the build finishes, it will output the IDs of the new AMIs. To see how to deploy one of these AMIs, check out the 
[consul-cluster example](https://github.com/hashicorp/terraform-aws-consul/tree/master/MAIN.md).




## Creating your own Packer template for production usage

When creating your own Packer template for production usage, you can copy the example in this folder more or less 
exactly, except for one change: we recommend replacing the `file` provisioner with a call to `git clone` in the `shell` 
provisioner. Instead of:

```json
{
  "provisioners": [{
    "type": "file",
    "source": "{{template_dir}}/../../../terraform-aws-consul",
    "destination": "/tmp"
  },{
    "type": "shell",
    "inline": [
      "/tmp/terraform-aws-consul/modules/install-consul/install-consul --version {{user `consul_version`}}",
      "/tmp/terraform-aws-consul/modules/install-dnsmasq/install-dnsmasq"
    ],
    "pause_before": "30s"
  }]
}
```

Your code should look more like this:

```json
{
  "provisioners": [{
    "type": "shell",
    "inline": [
      "git clone --branch <MODULE_VERSION> https://github.com/hashicorp/terraform-aws-consul.git /tmp/terraform-aws-consul",
      "/tmp/terraform-aws-consul/modules/install-consul/install-consul --version {{user `consul_version`}}",
      "/tmp/terraform-aws-consul/modules/install-dnsmasq/install-dnsmasq"
    ],
    "pause_before": "30s"
  }]
}
```

You should replace `<MODULE_VERSION>` in the code above with the version of this module that you want to use (see
the [Releases Page](../../releases) for all available versions). That's because for production usage, you should always
use a fixed, known version of this Module, downloaded from the official Git repo. On the other hand, when you're 
just experimenting with the Module, it's OK to use a local checkout of the Module, uploaded from your own 
computer.