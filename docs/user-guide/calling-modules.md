# Calling Modules

You can inspect not only the root module but also [module calls](https://developer.hashicorp.com/terraform/language/modules/syntax#calling-a-child-module). TFLint evaluates each call (i.e. `module` block) and emits any issues that result from the specified input variables.

```hcl
module "aws_instance" {
  source        = "./module"

  ami           = "ami-b73b63a0"
  instance_type = "t1.2xlarge"
}
```

```console
$ tflint
1 issue(s) found:

Error: instance_type is not a valid value (aws_instance_invalid_type)

  on template.tf line 5:
   5:   instance_type = "t1.2xlarge"

Callers:
   template.tf:5,19-31
   module/instance.tf:5,19-36

```

By default, TFLint only calls local modules whose the `source` is a relative path like `./*`. If you want to call remote modules (registry, git, etc.), you must run `terraform init` (or `terraform get`) before invoking TFLint so that modules are loaded into the `.terraform` directory. After that, invoke TFLint with `--call-module-type=all`.

```console
$ terraform init
$ tflint --call-module-type=all
```

The `--call-module-type` can also be set via configuration:

```hcl
config {
  call_module_type = "all"
}
```

If you don't want to call any modules, pass `--call-module-type=none`:

```console
$ tflint --call-module-type=none
```

If you want to ignore inspection for a particular module, you can use `--ignore-module`:

```console
$ tflint --ignore-module=./module
```

## Caveats

* Issues _must_ be associated with a variable that was passed to the module. If an issue within a child module is detected in an expression that does not reference a variable (`var`), it will be discarded.
* Rules that evaluate syntax rather than content _should_ ignore child modules.
* If you want to evaluate all TFLint rules on non-root modules, inspect directly against the module directories. Note that there is a difference between calling a child module in an inspection and inspecting a child module as the root module.
