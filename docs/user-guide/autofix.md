# Autofix

Some issues reported by TFLint can be auto-fixable. Auto-fixable issues are marked as "Fixable" as follows:

```console
$ tflint
1 issue(s) found:

Warning: [Fixable] Single line comments should begin with # (terraform_comment_syntax)

  on main.tf line 1:
   1: // locals values
   2: locals {

```

When run with the `--fix` option, TFLint will fix issues automatically.

```console
$ tflint --fix
1 issue(s) found:

Warning: [Fixed] Single line comments should begin with # (terraform_comment_syntax)

  on main.tf line 1:
   1: // locals values
   2: locals {

```

Please note that not all issues are fixable. The rule must support autofix.

If autofix is applied, it will automatically format the entire file. As a result, unrelated ranges may change.
