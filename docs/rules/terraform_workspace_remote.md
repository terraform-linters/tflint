# terraform_workspace_remote

`terraform.workspace` should not be used with a "remote" backend with remote execution. 

If remote operations are [disabled](https://www.terraform.io/docs/cloud/run/index.html#disabling-remote-operations) for your workspace, you can safely disable this rule:

```hcl
rule "terraform_workspace_remote" {
  enabled = false
}
```

## Example

```hcl
terraform {
  backend "remote" {
    # ...
  }
}

resource "aws_instance" "a" {
  tags = {
    workspace = terraform.workspace
  }
}
```

```
$ tflint
1 issue(s) found:

Warning: terraform.workspace should not be used with a 'remote' backend (terraform_workspace_remote)

  on example.tf line 8:
   8:   tags = {
   9:     workspace = terraform.workspace
  10:   }

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.5/docs/rules/terraform_workspace_remote.md
```

## Why

Terraform configuration may include the name of the [current workspace](https://www.terraform.io/docs/state/workspaces.html#current-workspace-interpolation) using the `${terraform.workspace}` interpolation sequence. However, when Terraform Cloud workspaces are executing Terraform runs remotely, the Terraform CLI always uses the `default` workspace.

The [remote](https://www.terraform.io/docs/backends/types/remote.html) backend is used with Terraform Cloud workspaces. Even if you set a `prefix` in the `workspaces` block, this value will be ignored during remote runs.

For more information, see the [`remote` backend workspaces documentation](https://www.terraform.io/docs/backends/types/remote.html#workspaces).

## How To Fix

Consider adding a variable to your configuration and setting it in each cloud workspace:

```tf
variable "workspace" {
  type        = string
  description = "The workspace name" 
}
```

You can also name the variable based on what the workspace suffix represents in your configuration (e.g. environment).
