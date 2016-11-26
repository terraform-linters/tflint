# AWS Instance Previous Type
Report this issue if you have specified the previous instance type. This issue type is WARNING.

## Example
```
resource "aws_instance" "web" {
  ami                  = "ami-b73b63a0"
  instance_type        = "t1.micro" # previous type!
  iam_instance_profile = "app-service"

  tags {
    Name = "HelloWorld"
  }
}
```

The following is the execution result of TFLint: 

```
$ tflint
template.tf
        WARNING:3 "t1.micro" is previous generation instance type.

Result: 1 issues  (0 errors , 1 warnings , 0 notices)
```

## Why
There are two types of instance types, the current generation and the previous generation. The current generation is superior to the previous generation in terms of performance and fee. AWS also officially states that unless there is a special reason, you should use the instance type of the current generation.

## How to fix
Follow the [upgrade paths](https://aws.amazon.com/ec2/previous-generation/) and confirm that the instance type of the current generation can be used, then select again.
