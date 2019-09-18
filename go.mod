module github.com/wata727/tflint

go 1.13

require (
	github.com/aws/aws-sdk-go v1.23.15
	github.com/fatih/color v1.7.0
	github.com/golang/mock v1.3.1
	github.com/google/go-cmp v0.3.1
	github.com/hashicorp/aws-sdk-go-base v0.3.0
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/hcl v0.0.0-20180404174102-ef8a98b0bbce // indirect
	github.com/hashicorp/hcl2 v0.0.0-20190821123243-0c888d1241f6
	github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/terraform v0.12.9
	github.com/jessevdk/go-flags v1.4.0
	github.com/mattn/go-colorable v0.1.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/afero v1.2.2
	github.com/zclconf/go-cty v1.0.1-0.20190708163926-19588f92a98f
)

// Override since git.apache.org is down
replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
