# Annotations

Annotation comments can disable rules on specific lines:

```hcl
resource "aws_instance" "foo" {
    # tflint-ignore: aws_instance_invalid_type
    instance_type = "t1.2xlarge"
}
```

The annotation works only for the same line or the line below it. You can also use `tflint-ignore: all` if you want to ignore all the rules.
