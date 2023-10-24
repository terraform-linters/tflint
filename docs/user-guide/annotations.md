# Annotations

Annotation comments can disable rules on specific lines:

```hcl
resource "aws_instance" "foo" {
    # tflint-ignore: aws_instance_invalid_type
    instance_type = "t1.2xlarge"
}
```

Multiple rules can be specified as a comma-separated list:

```hcl
resource "aws_instance" "foo" {
    # tflint-ignore: aws_instance_invalid_type, other_rule
    instance_type = "t1.2xlarge"
}
```

All rules can be ignored by specifying the `all` keyword:

```hcl
resource "aws_instance" "foo" {
    # tflint-ignore: all
    instance_type = "t1.2xlarge"
}
```

It is possible to add a reason for why a rule is ignored:

```hcl
resource "aws_instance" "foo" {
    # tflint-ignore: aws_instance_invalid_type # A good reason
    instance_type = "t1.2xlarge"
}
```
