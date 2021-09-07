# Module Inspection

By default, TFLint inspects only the root module. It can optionally also inspect [module calls](https://www.terraform.io/docs/configuration/blocks/modules/syntax.html#calling-a-child-module). When this option is enabled, TFLint evaluates each call (i.e. `module` block) and emits any issues that result from the specified input variables. Module inspection is designed to run on all module calls, whether the `source` is local (`./*`) or remote (registry, git, etc.). 

```hcl
module "aws_instance" {
  source        = "./module"

  ami           = "ami-b73b63a0"
  instance_type = "t1.2xlarge"
}
```

```console
$ tflint --module
1 issue(s) found:

Error: instance_type is not a valid value (aws_instance_invalid_type)

  on template.tf line 5:
   5:   instance_type = "t1.2xlarge"

Callers:
   template.tf:5,19-31
   module/instance.tf:5,19-36

```

## Caveats

* Module inspection mode _does not recursively search_ for Terraform modules. It follows `module` blocks in the root module where TFLint was invoked.
* Issues _must_ be associated with a variable that was passed to the module. If an issue within a child module is detected in an expression that does not reference a variable (`var`), it will be discarded.
* Rules that evaluate syntax rather than content _should_ ignore child modules.

If you want to evaluate all TFLint rules on non-root modules, pass their paths directly to TFLint.

## Enabling

Module inspection is disabled by default and can be enabled with the `--module` flag. It can also be enabled via configuration:

```hcl
config {
  module = true
}
```

You must run `terraform init` before invoking TFLint with module inspection so that modules are loaded into the `.terraform` directory. You can use the `--ignore-module` option if you want to skip inspection for a particular module:

```
tflint --ignore-module=./module
```
