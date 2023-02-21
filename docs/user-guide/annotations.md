# Annotations

Annotation comments can disable rules on specific lines:

```hcl
# tflint-ignore: aws_instance_invalid_type
resource "aws_instance" "foo" {
    instance_type = "t1.2xlarge"
}
```

Multiple rules can be specified as a comma-separated list:

```hcl
# tflint-ignore: aws_instance_invalid_type, other_rule
resource "aws_instance" "foo" {
    instance_type = "t1.2xlarge"
}
```

All rules can be ignored by specifying the `all` keyword:

```hcl
# tflint-ignore: all
resource "aws_instance" "foo" {
    instance_type = "t1.2xlarge"
}
```
