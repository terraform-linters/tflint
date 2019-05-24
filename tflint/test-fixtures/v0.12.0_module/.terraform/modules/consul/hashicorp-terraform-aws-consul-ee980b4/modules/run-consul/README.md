# Consul Run Script

This folder contains a script for configuring and running Consul on an [AWS](https://aws.amazon.com/) server. This 
script has been tested on the following operating systems:

* Ubuntu 16.04
* Amazon Linux

There is a good chance it will work on other flavors of Debian, CentOS, and RHEL as well.




## Quick start

This script assumes you installed it, plus all of its dependencies (including Consul itself), using the [install-consul 
module](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/install-consul). The default install path is `/opt/consul/bin`, so to start Consul in server mode, 
you run:

```
/opt/consul/bin/run-consul --server
```

To start Consul in client mode, you run:
 
```
/opt/consul/bin/run-consul --client
```

This will:

1. Generate a Consul configuration file called `default.json` in the Consul config dir (default: `/opt/consul/config`).
   See [Consul configuration](#consul-configuration) for details on what this configuration file will contain and how
   to override it with your own configuration.
   
1. Generate a [Supervisor](http://supervisord.org/) configuration file called `run-consul.conf` in the Supervisor
   config dir (default: `/etc/supervisor/conf.d`) with a command that will run Consul:  
   `consul agent -config-dir=/opt/consul/config -data-dir=/opt/consul/data`.

1. Tell Supervisor to load the new configuration file, thereby starting Consul.

We recommend using the `run-consul` command as part of [User 
Data](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html#user-data-shell-scripts), so that it executes
when the EC2 Instance is first booting. After runing `run-consul` on that initial boot, the `supervisord` configuration 
will automatically restart Consul if it crashes or the EC2 instance reboots.

See the [consul-cluster example](https://github.com/hashicorp/terraform-aws-consul/tree/master/MAIN.md) for fully-working sample code.




## Command line Arguments

The `run-consul` script accepts the following arguments:

* `server` (optional): If set, run in server mode. Exactly one of `--server` or `--client` must be set.
* `client` (optional): If set, run in client mode. Exactly one of `--server` or `--client` must be set. 
* `cluster-tag-key` (optional): Automatically form a cluster with Instances that have this tag key and the tag value
  in `--cluster-tag-value`.
* `cluster-tag-value` (optional): Automatically form a cluster with Instances that have the tag key in 
  `--cluster-tag-key` and this tag value.
* `config-dir` (optional): The path to the Consul config folder. Default is to take the absolute path of `../config`, 
  relative to the `run-consul` script itself.
* `data-dir` (optional): The path to the Consul config folder. Default is to take the absolute path of `../data`, 
  relative to the `run-consul` script itself.
* `user` (optional): The user to run Consul as. Default is to use the owner of `config-dir`.
* `skip-consul-config` (optional): If this flag is set, don't generate a Consul configuration file. This is useful if 
  you have a custom configuration file and don't want to use any of of the default settings from `run-consul`. 

Example:

```
/opt/consul/bin/run-consul --server --cluster-tag-key consul-cluster --cluster-tag-value prod-cluster 
```




## Consul configuration

`run-consul` generates a configuration file for Consul called `default.json` that tries to figure out reasonable 
defaults for a Consul cluster in AWS. Check out the [Consul Configuration Files 
documentation](https://www.consul.io/docs/agent/options.html#configuration-files) for what configuration settings are
available.
  
  
### Default configuration

`run-consul` sets the following configuration values by default:
  
* [advertise_addr](https://www.consul.io/docs/agent/options.html#advertise_addr): Set to the EC2 Instance's private IP 
  address, as fetched from [Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).

* [bind_addr](https://www.consul.io/docs/agent/options.html#bind_addr): Set to the EC2 Instance's private IP address, 
  as fetched from [Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).

* [bootstrap_expect](https://www.consul.io/docs/agent/options.html#bootstrap_expect): If `--server` is set, 
  set this config based on the EC2 Instance's tags (using the 
  [describe-tags API](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-tags.html)): 
    * If there is a `aws:autoscaling:groupName` tag, that means this EC2 Instance is part of an Auto Scaling Group 
      (ASG), so set this config to the desired capacity of the ASG (fetched via the [describe-auto-scaling-groups 
      API](https://docs.aws.amazon.com/cli/latest/reference/autoscaling/describe-auto-scaling-groups.html)). 
    * Otherwise, log a warning, and set this to 1. This fallback is not recommended!     

* [client_addr](https://www.consul.io/docs/agent/options.html#client_addr): Set to 0.0.0.0 so you can access the client
  and UI endpoint on each EC2 Instance from the outside.

* [datacenter](https://www.consul.io/docs/agent/options.html#datacenter): Set to the current AWS region (e.g. 
  `us-east-1`), as fetched from [Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).

* [node_name](https://www.consul.io/docs/agent/options.html#node_name): Set to the instance id, as fetched from 
  [Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).

* [retry_join_ec2](https://www.consul.io/docs/agent/options.html#retry_join_ec2): Look up the EC2 Instances tags
  (using the [describe-tags API](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-tags.html)) and set the
  following keys for this setting:
    * [tag_key](https://www.consul.io/docs/agent/options.html#tag_key): Set to the value of the `--cluster-tag-key`
      argument.
    * [tag_value](https://www.consul.io/docs/agent/options.html#tag_value): Set to the value this EC2 Instance has for
      the `tag_key`. If the key is not set, then the `retry_join_ec2` setting will NOT be included in the config file.
    * [region](https://www.consul.io/docs/agent/options.html#region): Set to the current AWS region (e.g. `us-east-1`), 
      as fetched from [Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).
      
* [server](https://www.consul.io/docs/agent/options.html#server): Set to true if `--server` is set.

* [ui](https://www.consul.io/docs/agent/options.html#ui): Set to true.


### Overriding the configuration

To override the default configuration, simply put your own configuration file in the Consul config folder (default: 
`/opt/consul/config`), but with a name that comes later in the alphabet than `default.json` (e.g. 
`my-custom-config.json`). Consul will load all the `.json` configuration files in the config dir and 
[merge them together in alphabetical order](https://www.consul.io/docs/agent/options.html#_config_dir), so that 
settings in files that come later in the alphabet will override the earlier ones. 

For example, to override the default `retry_join_ec2` settings, you could create a file called `tags.json` with the
contents:

```json
{
  "retry_join_ec2": {
    "tag_key": "custom-key",
    "tag_value": "custom-value",
    "region": "us-west-1"
  }
}
```

If you want to override *all* the default settings, you can tell `run-consul` not to generate a default config file
at all using the `--skip-consul-config` flag:

```
/opt/consul/bin/run-consul --server --skip-consul-config
```


### Required permissions

The `run-consul` script assumes it is running on an EC2 Instance with an [IAM 
Role](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html) that has the following permissions:

* `ec2:DescribeInstances`
* `ec2:DescribeTags`
* `autoscaling:DescribeAutoScalingGroups`

These permissions are automatically added by the [consul-cluster module](https://github.com/hashicorp/terraform-aws-consul/tree/master/modules/consul-cluster).




## How do you handle encryption?

Consul can encrypt all of its network traffic (see the [encryption docs for 
details](https://www.consul.io/docs/agent/encryption.html)), but by default, encryption is not enabled in this 
Module. To enable encryption, you need to do the following:

1. [Gossip encryption: provide an encryption key](#gossip-encryption-provide-an-encryption-key)
1. [RPC encryption: provide TLS certificates](#rpc-encryption-provide-tls-certificates)


### Gossip encryption: provide an encryption key

To enable Gossip encryption, you need to provide a 16-byte, Base64-encoded encryption key, which you can generate using
the [consul keygen command](https://www.consul.io/docs/commands/keygen.html). You can put the key in a Consul 
configuration file (e.g. `encryption.json`) in the Consul config dir (default location: `/opt/consul/config`):

```json
{
  "encrypt": "cg8StVXbQJ0gPvMd9o7yrg=="
}
```


### RPC encryption: provide TLS certificates

To enable RPC encryption, you need to provide the paths to the CA and signing keys ([here is a tutorial on generating 
these keys](http://russellsimpkins.blogspot.com/2015/10/consul-adding-tls-using-self-signed.html)). You can specify 
these paths in a Consul configuration file (e.g. `encryption.json`) in the Consul config dir (default location: 
`/opt/consul/config`):

```json
{
  "ca_file": "/opt/consul/tls/certs/ca-bundle.crt",
  "cert_file": "/opt/consul/tls/certs/my.crt",
  "key_file": "/opt/consul/tls/private/my.key"
}
```

You will also want to set the [verify_incoming](https://www.consul.io/docs/agent/options.html#verify_incoming) and
[verify_outgoing](https://www.consul.io/docs/agent/options.html#verify_outgoing) settings to verify TLS certs on 
incoming and outgoing connections, respectively:

```json
{
  "ca_file": "/opt/consul/tls/certs/ca-bundle.crt",
  "cert_file": "/opt/consul/tls/certs/my.crt",
  "key_file": "/opt/consul/tls/private/my.key",
  "verify_incoming": true,
  "verify_outgoing": true
}
```



