# aws_resource_missing_tags

Require specific tags for all AWS resource types that support them.

## Configuration

```hcl
rule "aws_resource_missing_tags" {
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

Notice: aws_instance.instance is missing the following tags: "Bar", "Foo". (aws_resource_missing_tags)

  on test.tf line 3:
   3:   tags = {
   4:     foo = "Bar"
   5:     bar = "Baz"
   6:   }
```

## Why

You want to set a standardized set of tags for your AWS resources.

## How To Fix

For each resource type that supports tags, ensure that each missing tag is present.
