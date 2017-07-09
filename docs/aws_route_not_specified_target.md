# AWS Route Not Specified Target
Report this issue if routing target is not specified. This issue type is ERROR.

## Example
```
resource "aws_route" "foo" {
  route_table_id         = "rtb-1234abcd"
  destination_cidr_block = "10.0.1.0/22"
}
```

The following is the execution result of TFLint:


```
$ tflint
template.tf
        ERROR:1 The routing target is not specified, each routing must contain either a gateway_id, egress_only_gateway_id a nat_gateway_id, an instance_id or a vpc_peering_connection_id or a network_interface_id.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
`aws_route` creates new routing in route tables. If a routing target is not specified, an error will occur when run `terraform apply`.

## How To Fix
Please add a routing target. There are kinds of `gateway_id`, `egress_only_gateway_id`, `nat_gateway_id`, `instance_id`, `vpc_peering_connection_id`, `network_interface_id`.
