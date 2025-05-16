# Annotations

TFLint supports several comment annotations for suppressing issues for specific lines or files. Annotations can only suppress _issues_ emitted from fully valid, parseable Terraform modules. _Errors_ cannot be ignored.

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

It's a good idea to add a reason for why a rule is ignored, especially temporarily:

```hcl
resource "aws_instance" "foo" {
    # This instance type is new and TFLint doesn't know about it yet
    # tflint-ignore: aws_instance_invalid_type
    instance_type = "t10.2xlarge"
}
```

Or, on the same line:

```hcl
resource "aws_instance" "foo" {
  # tflint-ignore: aws_instance_invalid_type # too new for TFLint
  instance_type = "t10.2xlarge" 
}
```

The `//` comment style is also supported, but Terraform recommends `#`.

```hcl
resource "aws_instance" "foo" {
  // tflint-ignore: aws_instance_invalid_type // too new for TFLint
  instance_type = "t10.2xlarge" 
}
```

## Files

To disable an entire file, you can also use the `tflint-ignore-file` annotation:

```hcl
# tflint-ignore-file: aws_instance_invalid_type

resource "aws_instance" "foo" {
  instance_type = "t1.2xlarge"
}
```

This annotation is valid only at the top of the file. The following cannot be used and will result in an error:

```hcl
resource "aws_instance" "foo" {
  # tflint-ignore-file: aws_instance_invalid_type
  instance_type = "t1.2xlarge"
}
```

```hcl
resource "aws_instance" "foo" { # tflint-ignore-file: aws_instance_invalid_type
  instance_type = "t1.2xlarge"
}
```

## JSON

The `tflint-ignore-file` annotation is also supported in Terraform JSON by using a top-level [comment property](https://developer.hashicorp.com/terraform/language/syntax/json#comment-properties):

```json
{
  "//": "tflint-ignore-file: aws_instance_invalid_type",
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}
```

As with annotations in HCL files, multiple rules can be specified as a
comma-separated list:

```json
{
  "//": "tflint-ignore-file: aws_instance_invalid_type, other_rule",
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}
```

Similarly, annotations in JSON can be followed with arbitrary comments, but the annotation must be the first thing in the comment property string:

```json
{
  "//": "tflint-ignore-file: aws_instance_invalid_type # This instance type is new and TFLint doesn't know about it yet",
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}
```

The `tflint-ignore` annotation is not supported in JSON configuration.
