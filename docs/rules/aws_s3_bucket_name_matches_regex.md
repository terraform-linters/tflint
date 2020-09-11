# aws_s3_bucket_name_match_regex

Ensures all s3 bucket names match a defined regex

## Configuration

```hcl
rule "aws_s3_bucket_name_match_regex" {
  enabled = true
  regex = "^prod.*"
}
```

## Examples

Most resources use the `tags` attribute with simple `key`=`value` pairs:

```hcl
resource "aws_s3_bucket" "foo" {
  bucket = "foo.domain.com"
  acl    = "private"

  tags = {
    Name        = "foo"
    Environment = "prod"
  }
}
```

```sh
$ tflint
1 issue(s) found:

Error: Bucket name foo.domain.com does not match regex ^prod.*(aws_s3_bucket_name_match_regex)

  on ../infrastructure/infrastructure/shared-services/s3-buckets.tf line 2:
  2:   bucket = "foo.domain.com"
```

## Why

You would like a standardized naming convention for all s3 buckets

## How To Fix

Ensure bucket name matches regex
