module github.com/wata727/tflint/tools

go 1.13

require (
	github.com/hashicorp/hcl/v2 v2.0.0
	github.com/hashicorp/hcl2 v0.0.0-20190821123243-0c888d1241f6
	github.com/hashicorp/terraform v0.12.9
	github.com/serenize/snaker v0.0.0-20171204205717-a683aaf2d516
	github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20190926213315-2a971231c1ba
)

// Override since git.apache.org is down
replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
