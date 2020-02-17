# aws_resource_tags

Require specific tags for all AWS resource types that support them.

## Configuration

```hcl
rule "terraform_module_pinned_source" {
  enabled = true
  tags = ["Foo", "Bar"]
}
```

## Example

```hcl
resource "aws_instance" "instance" {
  instance_type = "m5.large"
  tags = {
    foo = "Bar"
    bar = "Baz"
  }
}
```

```
$ tflint
1 issue(s) found:

Error: Wanted tags: Bar,Foo, found: bar,foo (aws_resource_tags)

  on test.tf line 3:
   3:   tags = {
   4:     foo = "Bar"
   5:     bar = "Baz"
   6:   }
```

## Why

You want to set a standardized set of tags for your AWS resources.

## How To Fix

Set the tags according to the rule configuration.
