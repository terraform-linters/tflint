# aws_route_specified_muletiple_targets

Disallow routes that have multiple targets.

## Example

```hcl
resource "aws_route" "foo" {
  route_table_id         = "rtb-1234abcd"
  destination_cidr_block = "10.0.1.0/22"
  gateway_id             = "igw-1234abcd"
  egress_only_gateway_id = "eigw-1234abcd" # second routing target?
}
```

```console
$ tflint
1 issue(s) found:

Error: More than one routing target specified. It must be one. (aws_route_specified_multiple_targets)

  on template.tf line 1:
   1: resource "aws_route" "foo" {

Reference: https://github.com/wata727/tflint/blob/v0.11.0/docs/rules/aws_route_specified_multiple_targets.md

```

## Why

It occurs an error.

## How To Fix

Check if two or more of the following attributes are specified:

- gateway_id
- egress_only_gateway_id
- nat_gateway_id
- instance_id
- vpc_peering_connection_id
- network_interface_id
- transit_gateway_id
