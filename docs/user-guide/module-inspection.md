# Module Inspection

By default, TFLint inspects only the root module, so if a resource to be inspected is cut out into a module, it will be ignored from inspection targets.

To avoid such problems, TFLint can also inspect [Module Calls](https://www.terraform.io/docs/configuration/blocks/modules/syntax.html#calling-a-child-module). In this case, it checks based on the input variables passed to the calling module.

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

Module inspection is disabled by default. Inspection is enabled by running with the `--module` option. Note that you need to run `terraform init` first because of TFLint loads modules in the same way as Terraform. 

You can use the `--ignore-module` option if you want to skip inspection for a particular module. Note that you need to pass module sources rather than module ids for backward compatibility.

```
$ tflint --ignore-module=./module
```
