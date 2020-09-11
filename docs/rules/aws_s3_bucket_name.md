# aws_s3_bucket_name

Ensures all S3 bucket names match the specified naming rules.

## Configuration

```hcl
rule "aws_s3_bucket_name" {
  enabled = true
  regex = "[a-z\-]+"
  prefix = "my-org"
}
```

* `regex`: A Go regex that bucket names must match (string)
* `prefix`: A prefix that should be used for bucket names (string)

## Examples

```hcl
resource "aws_s3_bucket" "foo" {
  bucket = "foo"
}
```

```sh
$ tflint
1 issue(s) found:

Warning: Bucket name "foo" does not have prefix "my-org" (aws_s3_bucket_name)

  on main.tf line 2:
  2:   bucket = "foo"
```

## Why

Amazon S3 bucket names must be globally unique and have [restrictive naming rules](https://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html#bucketnamingrules).

* Prefixing bucket names with an organization name can help avoid naming conflicts
* You may wish to enforce other naming conventions (e.g., disallowing dots)

## How To Fix

Ensure the bucket name matches the specified rules.
