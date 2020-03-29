# aws_resource_missing_tags

Require specific tags for all AWS resource types that support them.

## Configuration

```hcl
rule "aws_resource_missing_tags" {
  enabled = true
  tags = ["Foo", "Bar"]
  exclude = ["aws_autoscaling_group"] # (Optional) Exclude some resource types from tag checks
}
```

## Examples

Most resources use the `tags` attribute with simple `key`=`value` pairs:

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

Notice: aws_instance.instance is missing the following tags: "Bar", "Foo". (aws_resource_missing_tags)

  on test.tf line 3:
   3:   tags = {
   4:     foo = "Bar"
   5:     bar = "Baz"
   6:   }
```

Iterators in `dynamic` blocks cannot be expanded, so the tags in the following example will not be detected.

```hcl
locals {
  tags = [
    {
      key   = "Name",
      value = "SomeName",
    },
    {
      key   = "env",
      value = "SomeEnv",
    },
  ]
}
resource "aws_autoscaling_group" "this" {
  dynamic "tag" {
    for_each = local.tags

    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = true
    }
  }
}
```

## Why

You want to set a standardized set of tags for your AWS resources.

## How To Fix

For each resource type that supports tags, ensure that each missing tag is present.
