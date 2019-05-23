# Consul Install Script

This folder contains a script for installing Consul and its dependencies. Use this script along with the
[run-consul script](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/run-consul) to create a Consul [Amazon Machine Image 
(AMI)](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) that can be deployed in 
[AWS](https://aws.amazon.com/) across an Auto Scaling Group using the [consul-cluster module](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/consul-cluster).

This script has been tested on the following operating systems:

* Ubuntu 16.04
* Amazon Linux

There is a good chance it will work on other flavors of Debian, CentOS, and RHEL as well.



## Quick start

<!-- TODO: update the clone URL to the final URL when this Module is released -->

To install Consul, use `git` to clone this repository at a specific tag (see the [releases page](../../../../releases) 
for all available tags) and run the `install-consul` script:

```
git clone --branch <VERSION> https://github.com/hashicorp/terraform-aws-consul.git
terraform-aws-consul/modules/install-consul/install-consul --version 0.8.0
```

The `install-consul` script will install Consul, its dependencies, and the [run-consul script](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/run-consul).
The `run-consul` script is also run when the server is booting to start Consul and configure it to automatically 
join other nodes to form a cluster.

We recommend running the `install-consul` script as part of a [Packer](https://www.packer.io/) template to create a
Consul [Amazon Machine Image (AMI)](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) (see the 
[consul-ami example](https://github.com/hashicorp/terraform-aws-consul/tree/master/examples/consul-ami) for a fully-working sample code). You can then deploy the AMI across an Auto 
Scaling Group using the [consul-cluster module](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/consul-cluster) (see the [main 
example](https://github.com/hashicorp/terraform-aws-consul/tree/master/MAIN.md) for fully-working sample code).




## Command line Arguments

The `install-consul` script accepts the following arguments:

* `version VERSION`: Install Consul version VERSION. Required. 
* `path DIR`: Install Consul into folder DIR. Optional.
* `user USER`: The install dirs will be owned by user USER. Optional.

Example:

```
install-consul --version 0.8.0
```



## How it works

The `install-consul` script does the following:

1. [Create a user and folders for Consul](#create-a-user-and-folders-for-consul)
1. [Install Consul binaries and scripts](#install-consul-binaries-and-scripts)
1. [Install supervisord](#install-supervisord)
1. [Follow-up tasks](#follow-up-tasks)


### Create a user and folders for Consul

Create an OS user named `consul`. Create the following folders, all owned by user `consul`:

* `/opt/consul`: base directory for Consul data (configurable via the `--path` argument).
* `/opt/consul/bin`: directory for Consul binaries.
* `/opt/consul/data`: directory where the Consul agent can store state.
* `/opt/consul/config`: directory where the Consul agent looks up configuration.
* `/opt/consul/log`: directory where Consul will store log output.


### Install Consul binaries and scripts

Install the following:

* `consul`: Download the Consul zip file from the [downloads page](https://www.consul.io/downloads.html) (the version 
  number is configurable via the `--version` argument), and extract the `consul` binary into `/opt/consul/bin`. Add a
  symlink to the `consul` binary in `/usr/local/bin`.
* `run-consul`: Copy the [run-consul script](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/run-consul) into `/opt/consul/bin`. 


### Install supervisord

Install [supervisord](http://supervisord.org/). We use it as a cross-platform supervisor to ensure Consul is started
whenever the system boots and restarted if the Consul process crashes.


### Follow-up tasks

After the `install-consul` script finishes running, you may wish to do the following:

1. If you have custom Consul config (`.json`) files, you may want to copy them into the config directory (default:
   `/opt/consul/config`).
1. If `/usr/local/bin` isn't already part of `PATH`, you should add it so you can run the `consul` command without
   specifying the full path.
   


## Why use Git to install this code?

We needed an easy way to install these scripts that satisfied a number of requirements, including working on a variety 
of operating systems and supported versioning. Our current solution is to use `git`, but this may change in the future.
See [Package Managers](https://github.com/hashicorp/terraform-aws-consul/tree/master/_docs/package-managers.md) for a full discussion of the requirements, trade-offs, and why we
picked `git`.
