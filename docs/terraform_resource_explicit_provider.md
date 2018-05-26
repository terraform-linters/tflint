# Terraform Resource Explicit Provider
This issue is reported if you don't set `provider` in resources explicitly. This rule is disabled by default.

## Example
```hcl
provider "aws" {
  alias  = "west"
  region = "us-west-1"
}

resource "aws_instance" "web" {
  ami           = "ami-b73b63a0"
  instance_type = "t2.micro"

  tags {
    Name = "HelloWorld"
  }
}
```

The following is execution result of TFLint:

```
$ tflint
template.tf
        WARNING:1 Resource "web" provider is implicit (terraform_resource_explicit_provider)

Result: 1 issues  (0 errors , 1 warnings , 0 notices)
```

## Why

Resources are normally associated with the default provider configuration inferred from the resource type name.
However, If you use multiple providers, you must explicitly specify the provider on each resource if you do not prepare defaults.

## How To Fix

Specify `provider` in the resource explicitly.

```hcl
provider "aws" {
  alias  = "west"
  region = "us-west-1"
}

resource "aws_instance" "web" {
  provider = "aws.west"

  ami           = "ami-b73b63a0"
  instance_type = "t2.micro"

  tags {
    Name = "HelloWorld"
  }
}
```
