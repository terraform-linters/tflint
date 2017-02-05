# AWS Instance Not Specified IAM Profile
Report this issue if IAM profile is not specified. This issue type is NOTICE.

## Example
```
resource "aws_instance" "web" {
  ami           = "ami-b73b63a0"
  instance_type = "m4.2xlarge"

  tags {
    Name = "HelloWorld"
  }
}
```

The following is the execution result of TFLint: 

```
$ tflint
template.tf
        NOTICE:1 "iam_instance_profile" is not specified. If you want to change it, you need to recreate it

Result: 1 issues  (0 errors , 0 warnings , 1 notices)
```

## Why
You can select only one IAM profile at instance setup. However, if you do not select it, you will need to recreate the instance if you want to select it. The mechanism of IAM profile is the method for handling credentials most securely on AWS and in many cases it is necessary when you want to use another service on AWS (for example, S3).

Even if you think that you do not need an IAM profile, we recommend that you specify a dummy. Then you can change the privilege when you need it, so you can escape the recreate of the instance.

## How To Fix
Please add `iam_instance_profile` attribute.
