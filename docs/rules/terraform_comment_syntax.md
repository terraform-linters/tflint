# terraform_comment_syntax

Disallow `//` comments in favor of `#`.

## Example

```hcl
# Good
// Bad
```

```
$ tflint
1 issue(s) found:

Warning: Single line comments should begin with # (terraform_comment_syntax)

  on main.tf line 2:
   2: // Bad

Reference: https://github.com/terraform-linters/tflint-ruleset-terraform/blob/v0.1.0/docs/rules/terraform_comment_syntax.md
```

## Why

The Terraform language supports two different syntaxes for single-line comments: `#` and `//`. However, `#` is the default comment style and should be used in most cases.

* [Configuration Syntax: Comments](https://www.terraform.io/docs/configuration/syntax.html#comments)

## How To Fix

Replace the leading double-slash (`//`) in your comment with the number sign (`#`).
