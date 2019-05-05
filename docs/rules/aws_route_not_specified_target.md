# aws_route_not_specified_target

Disallow routes that have no targets.

## Example

```hcl
resource "aws_route" "foo" {
  route_table_id         = "rtb-1234abcd"
  destination_cidr_block = "10.0.1.0/22"
}
```

```
$ tflint
template.tf
        ERROR:1 The routing target is not specified, each aws_route must contain either egress_only_gateway_id, gateway_id, instance_id, nat_gateway_id, network_interface_id, transit_gateway_id, or vpc_peering_connection_id.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why

It occurs an error.

## How To Fix

Add a routing target. There are [kinds of](https://www.terraform.io/docs/providers/aws/r/route.html#argument-reference) `egress_only_gateway_id`, `gateway_id`, `instance_id`, `nat_gateway_id`, `network_interface_id`, `transit_gateway_id`, `vpc_peering_connection_id`.
