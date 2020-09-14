# aws_s3_bucket_name

Ensures all s3 bucket names match a defined regex

## Configuration

```hcl
rule "aws_s3_bucket_name" {
  enabled = true
  regex = "^blue.*"
}
```

## Examples

Most resources use the `tags` attribute with simple `key`=`value` pairs:

```hcl
resource "aws_s3_bucket" "foo" {
  bucket = "foo.domain.com"
  acl    = "private"
}
```

```sh
$ tflint
1 issue(s) found:

Error: Bucket name foo.domain.com does not match regex ^blue.* (aws_s3_bucket_name)

  on ../infrastructure/infrastructure/shared-services/s3-buckets.tf line 2:
  2:   bucket = "foo.domain.com"
```

## Why

You would like a standardized naming convention for all s3 buckets

## How To Fix

Ensure bucket name matches regex
