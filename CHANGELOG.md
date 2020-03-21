## 0.15.3 (2020-03-21)

### Enhancements

- [#676](https://github.com/terraform-linters/tflint/pull/676): Bump terraform to v0.12.24 ([@chenrui333](https://github.com/chenrui333))
- [#681](https://github.com/terraform-linters/tflint/pull/681): Bump terraform-provider-aws to v2.54.0 from v2.53.0 ([@wata727](https://github.com/wata727))
- [#682](https://github.com/terraform-linters/tflint/pull/682): Bump github.com/aws/aws-sdk-go from 1.29.24 to 1.29.29

### BugFixes

- [#670](https://github.com/terraform-linters/tflint/pull/670): Adds support for aws_autoscaling_group tag blocks and tags attributes ([@bwhaley](https://github.com/bwhaley))
- [#679](https://github.com/terraform-linters/tflint/pull/679): Add bucket-owner-full-control to allowed S3 ACLs ([@sds](https://github.com/sds))
- [#680](https://github.com/terraform-linters/tflint/pull/680): Add bucket-owner-read to allowed S3 ACLs ([@wata727](https://github.com/wata727))

### Chores

- [#664](https://github.com/terraform-linters/tflint/pull/664): Use checkout action v2 ([@chenrui333](https://github.com/chenrui333))
- [#671](https://github.com/terraform-linters/tflint/pull/671): Bump github.com/golang/mock from 1.4.1 to 1.4.3
- [#675](https://github.com/terraform-linters/tflint/pull/675): Bump github.com/hashicorp/go-plugin from 1.1.0 to 1.2.0

## 0.15.2 (2020-03-14)

### Enhancements

- [#650](https://github.com/terraform-linters/tflint/pull/650): Bump github.com/zclconf/go-cty from 1.3.0 to 1.3.1
- [#653](https://github.com/terraform-linters/tflint/pull/653): Bump terraform to v0.12.23 ([@chenrui333](https://github.com/chenrui333))
- [#665](https://github.com/terraform-linters/tflint/pull/665): Bump github.com/aws/aws-sdk-go from 1.28.9 to 1.29.24
- [#667](https://github.com/terraform-linters/tflint/pull/667): Bump terraform-provider-aws to v2.53.0 from v2.50.0 ([@wata727](https://github.com/wata727))

### Chores

- [#647](https://github.com/terraform-linters/tflint/pull/647): Bump github.com/hashicorp/go-plugin from 1.0.1 to 1.1.0
- [#649](https://github.com/terraform-linters/tflint/pull/649): Bump golang to v1.14 ([@chenrui333](https://github.com/chenrui333))
- [#657](https://github.com/terraform-linters/tflint/pull/657): Run linters in GitHub Actions ([@wata727](https://github.com/wata727))
- [#660](https://github.com/terraform-linters/tflint/pull/660): Bump github.com/hashicorp/terraform-plugin-sdk from 1.7.0 to 1.8.0 in /tools
- [#666](https://github.com/terraform-linters/tflint/pull/666): Add install guide for Windows (choco install) ([@aaronsteers](https://github.com/aaronsteers))

## 0.15.1 (2020-03-02)

### BugFixes

- [#645](https://github.com/terraform-linters/tflint/pull/645): Emit an issue when there is no tags definition ([@wata727](https://github.com/wata727))

### Chores

- [#630](https://github.com/terraform-linters/tflint/pull/630): Bump github.com/zclconf/go-cty from 1.2.1 to 1.3.0
- [#640](https://github.com/terraform-linters/tflint/pull/640): Bump github.com/golang/mock from 1.4.0 to 1.4.1 
- [#641](https://github.com/terraform-linters/tflint/pull/641): Bump github.com/mattn/go-colorable from 0.1.4 to 0.1.6

## 0.15.0 (2020-02-25)

This release introduces advanced rule configuration syntax. This allows you to customize each rule with its own options. At the moment, only the `terraform_module_pinned_source` rule has its own options. See [documentation](https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md#configuration) for details.

### Breaking Changes

- [#624](https://github.com/terraform-linters/tflint/pull/624): Merge terraform_module_semver_source into terraform_module_pinned_source ([@wata727](https://github.com/wata727))
  - The `terraform_module_semver_source` rule has been removed. Instead, use the `terraform_module_pinned_source` rule with `semver` style option.

### Enhancements

- [#609](https://github.com/terraform-linters/tflint/pull/609): Add new terraform_deprecated_interpolation rule ([@wata727](https://github.com/wata727))
- [#619](https://github.com/terraform-linters/tflint/pull/619): Show the actual value in messages ([@wata727](https://github.com/wata727))
- [#629](https://github.com/terraform-linters/tflint/pull/629): Bump terraform to v0.12.21 ([@chenrui333](https://github.com/chenrui333))
- [#617](https://github.com/terraform-linters/tflint/pull/617): Check for tags on AWS resources ([@bwhaley](https://github.com/bwhaley))
- [#635](https://github.com/terraform-linters/tflint/pull/635): Bump terraform-provider-aws to v2.50.0 from v2.46.0 ([@wata727](https://github.com/wata727))

### Chores

- [#607](https://github.com/terraform-linters/tflint/pull/607): Add FAQ ([@wata727](https://github.com/wata727))
- [#608](https://github.com/terraform-linters/tflint/pull/608): Remove go111module on env variable in Dockerfile ([@cedarkuo](https://github.com/cedarkuo))
- [#610](https://github.com/terraform-linters/tflint/pull/610): Add docker build actions ([@wata727](https://github.com/wata727))
- [#637](https://github.com/terraform-linters/tflint/pull/637): Regenerate tags rule ([@wata727](https://github.com/wata727))

## 0.14.0 (2020-01-31)

This release ships an experimental plugin system again! The new plugin system supports all operating systems and works perfectly correctly. See [the documentation](https://github.com/terraform-linters/tflint/blob/v0.14.0/docs/guides/extend.md) about how to use and create plugins.

### Enhancements

- [#568](https://github.com/terraform-linters/tflint/pull/568): Add new rule: terraform_dash_in_output_name ([@osulli](https://github.com/osulli))
- [#578](https://github.com/terraform-linters/tflint/pull/578): Bump github.com/fatih/color from 1.7.0 to 1.9.0
- [#579](https://github.com/terraform-linters/tflint/pull/579) [#597](https://github.com/terraform-linters/tflint/pull/597): Bump terraform to v0.12.20 ([@chenrui333](https://github.com/chenrui333))
- [#585](https://github.com/terraform-linters/tflint/pull/585): Introduce go-plugin based plugin system ([@wata727](https://github.com/wata727))
- [#601](https://github.com/terraform-linters/tflint/pull/601): Bump github.com/aws/aws-sdk-go from 1.26.8 to 1.28.9
- [#605](https://github.com/terraform-linters/tflint/pull/605): Bump terraform-provider-aws to v2.46.0 from v2.43.0 ([@wata727](https://github.com/wata727))

### BugFixes

- [#604](https://github.com/terraform-linters/tflint/pull/604): Prefer "ActiveMQ" over "ACTIVEMQ" as engine type ([@wata727](https://github.com/wata727))

### Chores

- [#519](https://github.com/terraform-linters/tflint/pull/519): Set up cache and artifact action ([@wata727](https://github.com/wata727))
- [#573](https://github.com/terraform-linters/tflint/pull/573): Bump github.com/hashicorp/hcl/v2 from 2.0.0 to 2.3.0 in /tools
- [#576](https://github.com/terraform-linters/tflint/pull/576): Bump github.com/google/go-cmp from 0.3.1 to 0.4.0
- [#583](https://github.com/terraform-linters/tflint/pull/583): Bump golang to v1.13.6 for Docker image ([@chenrui333](https://github.com/chenrui333))
- [#594](https://github.com/terraform-linters/tflint/pull/594): Bump github.com/golang/mock from 1.3.1 to 1.4.0
- [#603](https://github.com/terraform-linters/tflint/pull/603): Bump github.com/hashicorp/terraform-plugin-sdk from 1.4.1 to 1.6.0 in /tools

## 0.13.4 (2019-12-27)

### Enhancements

- [#563](https://github.com/terraform-linters/tflint/pull/563): Add elasticache support for t3 family ([@BrunoChauvet](https://github.com/BrunoChauvet))
- [#564](https://github.com/terraform-linters/tflint/pull/564): Bump github.com/aws/aws-sdk-go from 1.25.48 to 1.26.8
- [#565](https://github.com/terraform-linters/tflint/pull/565): Bump terraform-provider-aws to v2.43.0 from v2.41.0 ([@wata727](https://github.com/wata727))

### Chores

- [#566](https://github.com/terraform-linters/tflint/pull/566): Add GPG public key ([@wata727](https://github.com/wata727))

## 0.13.3 (2019-12-18)

### Enhancements

- [#545](https://github.com/terraform-linters/tflint/pull/545): Bump terraform to v0.12.18 ([@chenrui333](https://github.com/chenrui333))

### BugFixes

- [#555](https://github.com/terraform-linters/tflint/pull/555): Parse absolute paths in TF_DATA_DIR correctly ([@madddi](https://github.com/madddi))

### Chores

- [#542](https://github.com/terraform-linters/tflint/pull/542): Fix the pre-commit hook ([@Dunedan](https://github.com/Dunedan))
- [#556](https://github.com/terraform-linters/tflint/pull/556): Remove legacy TF 0.11 fields from module records ([@madddi](https://github.com/madddi))

## 0.13.2 (2019-12-07)

### Enhancements

- [#526](https://github.com/terraform-linters/tflint/pull/526) [#532](https://github.com/terraform-linters/tflint/pull/532): Bump terraform to v0.12.17 from v0.12.15 ([@chenrui333](https://github.com/chenrui333))
- [#537](https://github.com/terraform-linters/tflint/pull/537): Bump github.com/aws/aws-sdk-go from 1.25.31 to 1.25.48
- [#541](https://github.com/terraform-linters/tflint/pull/541): Bump terraform-provider-aws to v2.41.0 from v2.36.0 ([@wata727](https://github.com/wata727))

### Chores

- [#530](https://github.com/terraform-linters/tflint/pull/530): update the docker image name ([@ozbillwang](https://github.com/ozbillwang))
- [#534](https://github.com/terraform-linters/tflint/pull/534): Bump the base image to v1.13.5 ([@chenrui333](https://github.com/chenrui333))
- [#535](https://github.com/terraform-linters/tflint/pull/535): Pin actions/checkout@v1 ([@wata727](https://github.com/wata727))

## 0.13.1 (2019-11-16)

- [#524](https://github.com/terraform-linters/tflint/pull/524): Revert: Experimental plugin support ([@wata727](https://github.com/wata727))

## 0.13.0 (2019-11-16)

This is the first release in the terraform-linters organization. This release includes an experimental plugin system. You can easily add custom rules using the Go plugin system. Please see [here](https://github.com/terraform-linters/tflint/blob/v0.13.0/docs/guides/extend.md) for the detail.

### Breaking Changes

- [#496](https://github.com/terraform-linters/tflint/pull/496): Check invalid rule names ([@abitrolly](https://github.com/abitrolly))

### Enhancements

- [#500](https://github.com/terraform-linters/tflint/pull/500): Experimental plugin support ([@wata727](https://github.com/wata727))
- [#506](https://github.com/terraform-linters/tflint/pull/506) [#514](https://github.com/terraform-linters/tflint/pull/514): Bump github.com/aws/aws-sdk-go from 1.25.4 to 1.25.31 ([@chenrui333](https://github.com/chenrui333),[@wata727](https://github.com/wata727))
- [#506](https://github.com/terraform-linters/tflint/pull/506) [#523](https://github.com/terraform-linters/tflint/pull/523): Bump terraform to v0.12.15 from v0.12.10 ([@chenrui333](https://github.com/chenrui333),[@wata727](https://github.com/wata727))
- [#518](https://github.com/terraform-linters/tflint/pull/518): Add an optional checker for semver versions ([@alexwlchan](https://github.com/alexwlchan))
- [#522](https://github.com/terraform-linters/tflint/pull/522): Bump terraform-provider-aws from v2.32.0 to v2.36.0 ([@wata727](https://github.com/wata727))

### BugFixes

- [#517](https://github.com/terraform-linters/tflint/pull/517): When checking if a source is pinned, allow for Mercurial/Bitbucket ([@alexwlchan](https://github.com/alexwlchan))

### Chores

- [#488](https://github.com/terraform-linters/tflint/pull/488): Update base image to alpine v3.10 ([@chenrui333](https://github.com/chenrui333))
- [#503](https://github.com/terraform-linters/tflint/pull/503): add note about recursive check ([@IslamAzab](https://github.com/IslamAzab))
- [#515](https://github.com/terraform-linters/tflint/pull/515): Rename import path ([@wata727](https://github.com/wata727))
- [#516](https://github.com/terraform-linters/tflint/pull/516): Run tests on GitHub Actions ([@wata727](https://github.com/wata727))
- [#520](https://github.com/terraform-linters/tflint/pull/520): oneliner linux should follow redirects when fetching latest release ([@alexsn](https://github.com/alexsn))

## 0.12.1 (2019-10-12)

### Enhancements

- [#467](https://github.com/terraform-linters/tflint/pull/467): Bump github.com/mattn/go-colorable from 0.1.2 to 0.1.4
- [#476](https://github.com/terraform-linters/tflint/pull/476): Bump github.com/hashicorp/aws-sdk-go-base from 0.3.0 to 0.4.0
- [#482](https://github.com/terraform-linters/tflint/pull/482): TFLint is now compatible with Terraform v0.12.10
  - See https://github.com/hashicorp/terraform/releases/tag/v0.12.10
  - Support new built-in functions: `parseint` and `cidrsubnets`
- [#484](https://github.com/terraform-linters/tflint/pull/484): Bump terraform-provider-aws from v2.30.0 to v2.32.0

### Chores

- [#471](https://github.com/terraform-linters/tflint/pull/471): Bump TFLint version in issue template ([@abitrolly](https://github.com/abitrolly))
- [#474](https://github.com/terraform-linters/tflint/pull/474): Switch to HCL 2.0 in the HCL repository ([@explodingcamera](https://github.com/explodingcamera))
- [#487](https://github.com/terraform-linters/tflint/pull/487): Test tools in GitHub Actions

## 0.12.0 (2019-09-29)

This release includes an experimental Language Server Procotol support. Using LSP makes it easy to integrate TFLint with your favorite editor. Currently, only diagnostics are provided.

### Breaking Changes

- [#454](https://github.com/terraform-linters/tflint/pull/454): Remove deprecated `ignore-rule` option
  - `--ignore-rule` CLI flag and config attribute are removed. Please use `--disable-rule`, or define a `rule` block with `enabled = false` in your TFLint config file.
  - Note that `--disable-rule` behaves differently than `--ignore-rule`. Since `--ignore-rule` is deprecated, it was overridden by the value defined in rule blocks, but `--disable-rule` always takes precedence.

### Enhancements

- [#439](https://github.com/terraform-linters/tflint/pull/439): Experimental language server support
  - `tflint --langserver` launches a Language Server which speaks LSP v3.14.0.
- [#455](https://github.com/terraform-linters/tflint/pull/455): Add `--enable-rule` and `--disable-rule` options
- [#456](https://github.com/terraform-linters/tflint/pull/456): Allow specifying multiple `--ignore-module` and `--var-file` flags
  - You can use these flags multiple times. The previous style is still valid for backward compatibility.
- [#459](https://github.com/terraform-linters/tflint/pull/459): rule: Add m5, r5, and z1d RDS instance families and m3 and r3 families will be previous generations
- [#460](https://github.com/terraform-linters/tflint/pull/460): rule: Add m3 and r3 ElastiCache node types as previous generations
- [#461](https://github.com/terraform-linters/tflint/pull/461): rule: Add m3, c3, g2, r3, and i2 EC2 instance families as previous generations
- [#462](https://github.com/terraform-linters/tflint/pull/462): rule: Add aws-exec-read bucket ACL as a valid value
- [#463](https://github.com/terraform-linters/tflint/pull/463): Bump terraform-provider-aws from v2.28.1 to v2.30.0
  - Add g4dn instance family
  - The limit of length for config rule name will be changed 128 characters from 64
  - Add regexp validation for config rule name

### Chores

- [#449](https://github.com/terraform-linters/tflint/pull/449): docs: Add annotations page
- [#450](https://github.com/terraform-linters/tflint/pull/450): Add issue templates
- [#451](https://github.com/terraform-linters/tflint/pull/451): docs: Assume role is supported
- [#457](https://github.com/terraform-linters/tflint/pull/457): Tweak log levels
- [#458](https://github.com/terraform-linters/tflint/pull/458): Remove project package

## 0.11.2 (2019-09-19)

### Enhancements

- [#445](https://github.com/terraform-linters/tflint/pull/445): TFLint is now compatible with Terraform v0.12.9
  - See https://github.com/hashicorp/terraform/releases/tag/v0.12.9
- [#446](https://github.com/terraform-linters/tflint/pull/446): Bump terraform-provider-aws from v2.27.0 to v2.28.1
  - No changes for rules

### BugFixes

- [#442](https://github.com/terraform-linters/tflint/pull/442): aws_s3_bucket_invalid_region_rule: Fix false positives
- [#443](https://github.com/terraform-linters/tflint/pull/443): config: Fix panic when the expression includes invalid references

### Chores

- [#435](https://github.com/terraform-linters/tflint/pull/435): docs: Add Linix oneliner to download latest `tflint` ([@abitrolly](https://github.com/abitrolly))
- [#437](https://github.com/terraform-linters/tflint/pull/437): docs: Fix typo in dash rule ([@abitrolly](https://github.com/abitrolly))

## 0.11.1 (2019-09-11)

### Chores

- [#429](https://github.com/terraform-linters/tflint/pull/429) [#433](https://github.com/terraform-linters/tflint/pull/433): build: Upgrade to go 1.13 ([@chenrui333](https://github.com/chenrui333))
- [#431](https://github.com/terraform-linters/tflint/pull/431): build: Disable CGO in GoReleaser ([@craigfurman](https://github.com/craigfurman))

## 0.11.0 (2019-09-08)

This release includes major changes to the output format. In particular, third-party tool developers should be aware of changes to the JSON output format. Please see the "Breaking Changes" section for details.

### Breaking Changes

- [#396](https://github.com/terraform-linters/tflint/pull/396): Emit issues to the root module instead of each module
  - Previously issues found inside a module were reported along with the line number for that module, but it now reports on root module arguments that caused issues with the module.
- [#407](https://github.com/terraform-linters/tflint/pull/407): formatter: Multiple errors and context-rich pretty print
  - The output format of default and JSON has been changed. See the pull request for details.
- [#413](https://github.com/terraform-linters/tflint/pull/413): Remove `--quiet` option
  - This behavior is the default for new output formats.

### Enhancements

- [#395](https://github.com/terraform-linters/tflint/pull/395): config: Add support for `path.*` named values
- [#415](https://github.com/terraform-linters/tflint/pull/415): Add `--no-color` option
- [#421](https://github.com/terraform-linters/tflint/pull/421): Add mappings for new resources
  - 44 rules have been added.
- [#424](https://github.com/terraform-linters/tflint/pull/424): TFLint is now compatible with Terraform v0.12.8
  - See https://github.com/hashicorp/terraform/releases/tag/v0.12.8
- [#426](https://github.com/terraform-linters/tflint/pull/426): Bump terraform-provider-aws from v2.25.0 to v2.27.0
  - `aws_cur_report_definition_invalid_s3_region` rule now allows `ap-east-1` as a valid value.
  - `aws_instance_invalid_type`, `aws_launch_configuration_invalid_type` and `aws_launch_template_invalid_instance_type` rules now allow `i3en.metal` as a valid value.
  - `aws_ssm_parameter_invalid_tier` rule now allows `Intelligent-Tiering` as a valid value.
- [#423](https://github.com/terraform-linters/tflint/pull/423): client: Add support for role assumption
  - The `assume_role` block in the `provider` block is now taken into account.

### Chores

- [#410](https://github.com/terraform-linters/tflint/pull/410): Automatically generate API-based rules 
- [#411](https://github.com/terraform-linters/tflint/pull/411): Add tools task to Makefile and clean up
- [#412](https://github.com/terraform-linters/tflint/pull/412): docs: Tweak documentations
- [#414](https://github.com/terraform-linters/tflint/pull/414): docs: Fix exit status
- [#417](https://github.com/terraform-linters/tflint/pull/417): Refactoring tests
- [#419](https://github.com/terraform-linters/tflint/pull/419): Bump github.com/spf13/afero from 1.2.1 to 1.2.2
- [#428](https://github.com/terraform-linters/tflint/pull/428): Correct ineffassign ([@gliptak](https://github.com/gliptak))

## 0.10.3 (2019-08-24)

### Chores

- [#406](https://github.com/terraform-linters/tflint/pull/406): Remove GoReleaser before hooks

## 0.10.2 (2019-08-24)

### Enhancements

- [#404](https://github.com/terraform-linters/tflint/pull/404): Bump terraform-provider-aws from v2.24.0 to v2.25.0
  - No changes for rules.
- [#405](https://github.com/terraform-linters/tflint/pull/405): Bump terraform from v0.12.6 to v0.12.7
  - New functions `regex` and `regexall` are available.
  - See https://github.com/hashicorp/terraform/releases/tag/v0.12.7

### BugFixes

- [#400](https://github.com/terraform-linters/tflint/pull/400): rule: Fix values for excess_capacity_termination_policy. ([@alzabo](https://github.com/alzabo))

### Chores

- [#394](https://github.com/terraform-linters/tflint/pull/394): Remove image task from Makefile
- [#397](https://github.com/terraform-linters/tflint/pull/397): Bump github.com/hashicorp/terraform from 0.12.6 to 0.12.7 in /tools
- [#399](https://github.com/terraform-linters/tflint/pull/399): Release via GitHub Actions
- [#401](https://github.com/terraform-linters/tflint/pull/401): Manually maintain updated SDK-based validation rules

## 0.10.1 (2019-08-21)

### BugFixes

- [#393](https://github.com/terraform-linters/tflint/pull/393): Eval provider attributes
  - There is a bug that returned an error when using a variable in the `provider` block attributes.

## 0.10.0 (2019-08-17)

### Breaking Changes

- [#361](https://github.com/terraform-linters/tflint/pull/361): Get an AWS session in the same way as Terraform
  - It will take a region and access keys in the `provider` block written in configuration files into account.
  - Added support for ECS/CodeBuild task roles and EC2 roles.
  - There are breaking changes to credential priorities. It affects under the following cases:
    - If you have a region or access keys in the `provider` block, it prefers them over environment variables and shared credentials.
    - If there are environment variables and shared credentials, it prefers the environment variables. Previously, it prefers shared credentials.

### Changes

- [#378](https://github.com/terraform-linters/tflint/pull/378): Remove aws_instance_default_standard_volume rule
- [#379](https://github.com/terraform-linters/tflint/pull/379): Remove aws_db_instance_readable_password rule

### Enhancements

- [#384](https://github.com/terraform-linters/tflint/pull/384): Add terraform_dash_in_resource_name rule ([@kulinacs](https://github.com/kulinacs))
  - This rule is disabled by default.
- [#388](https://github.com/terraform-linters/tflint/pull/388): Bump terraform-provider-aws from v2.20.0 to v2.24.0
  - Added `me-south-1` as a valid region in `aws_route53_health_check_invalid_cloudwatch_alarm_region` rule and `aws_route53_zone_association_invalid_vpc_region` rule.
  - Added `capacityOptimized` as a valid strategy in `aws_spot_fleet_request_invalid_allocation_strategy` rule.

### Chores

- [#387](https://github.com/terraform-linters/tflint/pull/387): Bump github.com/google/go-cmp from 0.3.0 to 0.3.1
- [#389](https://github.com/terraform-linters/tflint/pull/389): Add Terraform compatibility badge
- [#390](https://github.com/terraform-linters/tflint/pull/390): Remove legacy module walkers

## 0.9.3 (2019-08-02)

### Enhancements

- [#375](https://github.com/terraform-linters/tflint/pull/375): Update dependencies to Terraform 0.12.6 ([@lawliet89](https://github.com/lawliet89))
  - Resource `for-each` syntax doesn't report an error, but TFLint still ignore `each.*` expressions.
  - See https://github.com/hashicorp/terraform/releases/tag/v0.12.6
- [#377](https://github.com/terraform-linters/tflint/pull/377): Bump terraform-provider-aws from v2.20.0 to v2.22.0
  - `aws_secretsmanager_secret_invalid_policy` rule now allows up to 20480.
  - `aws_secretsmanager_secret_version_invalid_secret_string` rule now allows up to 10240.
  - `aws_ssm_maintenance_window_target_invalid_resource_type` rule now allows `RESOURCE_GROUP` as a valid type.

### Chores

- [#368](https://github.com/terraform-linters/tflint/pull/368): Update brew instructions ([@arbourd](https://github.com/arbourd))
  - TFLint's formula is now hosted by `homebrew/core` 🎉
- [#373](https://github.com/terraform-linters/tflint/pull/373): Bump github.com/hashicorp/terraform from 0.12.5 to 0.12.6 in /tools

## 0.9.2 (2019-07-20)

### Enhancements

- [#360](https://github.com/terraform-linters/tflint/pull/360): Allow settings shared credentials file path
  - Added `--aws-creds-file` in CLI flags
  - Added `shared_credentials_file` in config attributes
- [#365](https://github.com/terraform-linters/tflint/pull/365): TFLint is now compatible with Terraform v0.12.5
  - See https://github.com/hashicorp/terraform/releases/tag/v0.12.4
  - See https://github.com/hashicorp/terraform/releases/tag/v0.12.5
- [#367](https://github.com/terraform-linters/tflint/pull/367): TFLint is now compatible with Terraform AWS provider v2.20.0
  - Updated `aws_cloudwatch_metric_alarm_invalid_comparison_operator` rule

## 0.9.1 (2019-07-09)

### Enhancements

- [#348](https://github.com/terraform-linters/tflint/pull/348): Update launch configuration instance types
- [#350](https://github.com/terraform-linters/tflint/pull/350): Add terraform_documented_variables/outputs rules
- [#356](https://github.com/terraform-linters/tflint/pull/356): Bump terraform-aws-provider from v2.16.0 to v2.18.0

### BugFixes

- [#355](https://github.com/terraform-linters/tflint/pull/355): Fix a false positive for `log-delivery-write` ACL

### Chores

- [#346](https://github.com/terraform-linters/tflint/pull/346): Docs: Limitations -> Compatibility with Terraform
- [#347](https://github.com/terraform-linters/tflint/pull/347): Fix rule generator

## 0.9.0 (2019-06-29)

This release includes breaking changes due to the removal of some CLI flags and options. Please see the "Breaking Changes" section for details.

As a major improvement, added 700+ rules in this release. These rules are automatically generated from aws-sdk validations and can be used without deep checking. For example, you can check whether a resource name matches the regular expression, whether it satisfies length constraints, whether it is included in the list of valid values, etc. before running `terraform plan` or `terraform apply`.

### Breaking Changes

- [#310](https://github.com/terraform-linters/tflint/pull/310): Remove `--fast` option
  - It disables only `aws_instance_invalid_ami` when passed this flag. But the rule is already faster in v0.8.2. Therefore, this flag is not necessary.
- [#311](https://github.com/terraform-linters/tflint/pull/311): Remove terraform_version option
  - `terraform_version` option is no longer used.
- [#313](https://github.com/terraform-linters/tflint/pull/313): Make non-zero exit status default if issues found
  - Previously, it has return 0 as exit status even if an issue was found, but now it will return 2.
  - If you would like to keep the previous behavior, you can use `--force` option.
- [#329](https://github.com/terraform-linters/tflint/pull/329): Disable module inspection by default
  - You no longer need to run `terraform init` just to run` tflint`.
  - If you also want to check module calls, pass the `--module` option. In that case, you need to run `terraform init` as before.

### Changes

- [#340](https://github.com/terraform-linters/tflint/pull/340): Replace aws_cloudwatch_metric_alarm_invalid_init with auto-generated
  - The output message has changed, but there has been no other change.

### Enhancements

- [#274](https://github.com/terraform-linters/tflint/pull/274): Auto generate rules from AWS API models
  - These rules are based on Terraform AWS provider v2.16.0.
- [#332](https://github.com/terraform-linters/tflint/pull/332), [#336](https://github.com/terraform-linters/tflint/pull/336): TFLint is now compatible with Terraform v0.12.3
  - See also https://github.com/hashicorp/terraform/releases/tag/v0.12.3
- [#343](https://github.com/terraform-linters/tflint/pull/343): Update valid instance type list

### BugFixes

- [#341](https://github.com/terraform-linters/tflint/pull/341): Fix false negatives in the S3 invalid ACL rule

### Chores

- [#326](https://github.com/terraform-linters/tflint/pull/326): Set up CI with Azure Pipelines
- [#337](https://github.com/terraform-linters/tflint/pull/337): Check mapping attribute types
- [#339](https://github.com/terraform-linters/tflint/pull/339): Remove appveyor.yml
- [#338](https://github.com/terraform-linters/tflint/pull/338): Mappings are checked based on Terraform v0.12.3 schema
- [#345](https://github.com/terraform-linters/tflint/pull/345): Revise documentations

## 0.8.3 (2019-06-09)

### Enhancements

- [#318](https://github.com/terraform-linters/tflint/pull/318): Added 3 checks for AWS Launch Configuration. ([@krzyzakp](https://github.com/krzyzakp))
  - `aws_launch_configuration_invalid_iam_profile`
  - `aws_launch_configuration_invalid_image_id`
  - `aws_launch_configuration_invalid_type`
- [#321](https://github.com/terraform-linters/tflint/pull/321): Add `--var` options.
- [#322](https://github.com/terraform-linters/tflint/pull/322): Add new rule: aws_s3_bucket_invalid_acl. ([@ineffyble](https://github.com/ineffyble))
- [#324](https://github.com/terraform-linters/tflint/pull/324): TFLint is now compatible with Terraform v0.12.1.
  - See also https://github.com/hashicorp/terraform/releases/tag/v0.12.1

### BugFixes

- [#320](https://github.com/terraform-linters/tflint/pull/320): Avoid InvalidAMIID errors.

### Others

- [#319](https://github.com/terraform-linters/tflint/pull/319): Added pre-commit hooks. ([@krzyzakp](https://github.com/krzyzakp))
- [#323](https://github.com/terraform-linters/tflint/pull/323): Bump github.com/aws/aws-sdk-go from 1.19.41 to 1.19.46

## 0.8.2 (2019-06-03)

### Enhancements

- [#308](https://github.com/terraform-linters/tflint/pull/308): Make aws_instance_invalid_ami rule faster.
  - The `--fast` option to disable this rule will be removed in v0.9.
- [#309](https://github.com/terraform-linters/tflint/pull/309): Accept a directory as an argument.

### Others

- [#298](https://github.com/terraform-linters/tflint/pull/298): Revise docker image.
- [#300](https://github.com/terraform-linters/tflint/pull/300): Bump github.com/mattn/go-colorable from 0.1.1 to 0.1.2.
- [#301](https://github.com/terraform-linters/tflint/pull/301): Bump github.com/mitchellh/go-homedir from 1.0.0 to 1.1.0.
- [#302](https://github.com/terraform-linters/tflint/pull/302): Bump github.com/aws/aws-sdk-go from 1.19.18 to 1.19.41.
- [#303](https://github.com/terraform-linters/tflint/pull/303): Bump github.com/k0kubun/pp from 2.3.0+incompatible to 2.4.0+incompatible.
- [#304](https://github.com/terraform-linters/tflint/pull/304): Bump github.com/hashicorp/go-version from 1.1.0 to 1.2.0.
- [#305](https://github.com/terraform-linters/tflint/pull/305): Bump github.com/golang/mock from 1.2.0 to 1.3.1.
- [#306](https://github.com/terraform-linters/tflint/pull/306): Bump github.com/google/go-cmp from 0.2.0 to 0.3.0.
- [#307](https://github.com/terraform-linters/tflint/pull/307): Remove mock package.

## 0.8.1 (2019-05-30)

### Enhancements

- [#277](https://github.com/terraform-linters/tflint/pull/277): Ignore annotation support.
  - `tflint-ignore: rule_name` annotation is now availble. See [README.md](https://github.com/terraform-linters/tflint/blob/v0.8.1/README.md#rules).

### BugFixes

- [#293](https://github.com/terraform-linters/tflint/pull/293): Fix false negatives when `aws_instance_default_standard_volume` rule checks `dynamic` blocks.
- [#297](https://github.com/terraform-linters/tflint/pull/297): Fix panic when checking whether an expression is null.

### Others

- [#292](https://github.com/terraform-linters/tflint/pull/292): Migrating to Go Modules.

## 0.8.0 (2019-05-25)

This release includes major changes due to being dependent on Terraform v0.12 internal API. While we try to keep backward compatibility as much as possible, it does include some breaking changes.

We strongly recommend [upgrading to Terraform v0.12](https://www.terraform.io/upgrade-guides/0-12.html) before trying TFLint v0.8. `terraform 0.12upgrade` is helpful to upgrade your configuration files.

### Breaking Changes

- Always return an error when failed to evaluate an expression.
  - Until now, except for module arguments, even if an error occurred, it was ignored.
  - Expressions including unsupported named values (such as `${module.foo}`) are not evaluated, so no error occurs.
- Drop support for `${terraform.env}`.
  - Previously `${terraform.env}` was a valid expression that returned the same as `${terraform.workspace}`.
  - This is because Terraform v0.12 doesn't support `${terraform.env}`.
- The file name of a module includes module ID instead of the source attribute.
  - Up to now it was output like `github.com/terraform-linters/example-module/instance.tf`, but it will be changed like `module_id/instance.tf`.
- Always parse all configuration files under the current directory.
  - When passing a file name as an argument, TFLint only parsed that file so far, but it now parses all configuration files under the current directory.
  - Also, file arguments are only used to filter the issues obtained. Therefore, you cannot pass files other than under the current directory.
  - As a known issue, If file arguments are passed, module's issues are not reported. This will be improved by changing handling of module's issues in the future.
  - These behaviors have been changed as it depends on Terraform's `configload` package.
  - In addition, modules are always loaded regardless of `ignore_module`.
- Raise an error when using invalid syntax as a Terraform configuration.
  - For example, it didn't raise an error when using `resources`(not `resource`) block because it is valid as HCL syntax in previous versions.
- Remove `--debug` option.
  - Please use `TFLINT_LOG` environment variables instead.
- Raise an error when a file passed by `--config` does not exist.
  - Previously the error was ignored and the default config was referenced.
- Remove duplicate resource rules.
  - This is due to technical difficulty and user experience.

### Enhancements

- HCL2 support
  - See also https://www.hashicorp.com/blog/terraform-0-1-2-preview
- Built-in Functions support
  - Until now, if an expression includes function calls, it was ignored.
- `TF_DATA_DIR` and `TF_WORKSPACE` environment variables are now available.
  - Until now, these variables are ignored.
- It is now possible to handle values doesn't have a default without raising errors.
  - In the past, an error occurred when there was a reference to a variable that had no default value in an attribute of a module. See [#205](https://github.com/terraform-linters/tflint/issues/205)
- Terraform v0.11 module support
  - Until now, it is failed to properly load a part of Terraform v0.11 module. See also [#167](https://github.com/terraform-linters/tflint/issues/167)
- Support for automatic loading `*.auto.tfvars` files.
  - Previously it was not loaded automatically.

### BugFixes

- Improve expression checks
  - Since it used to be checked by a regular expression, there were many bugs, but it was greatly improved by using the `terraform/lang` package. See [#204](https://github.com/terraform-linters/tflint/issues/204) [#160](https://github.com/terraform-linters/tflint/issues/160)
- Stop overwriting the config under the current directory by the config under the homedir.
  - Fixed the problem that overwrites the config under the current directory by homedir config.
- Improve to check for `aws_db_instance_readable_password`.
  - Previously, false positive occurred when setting values files or environment variables, but this problem has been fixed.
- Make `transit_gateway_id` as a valid target on `aws_route_specified_multiple_targets`

### Project Changes

- Change license: MIT -> MPL 2.0
  - See [#245](https://github.com/terraform-linters/tflint/pull/245)
- Update documentations
  - See [#272](https://github.com/terraform-linters/tflint/pull/272)

## 0.7.6 (2019-05-17)

### BugFixes

- [#276](https://github.com/terraform-linters/tflint/pull/276): Update aws_route_not_specified_target to handle transit_gateway_id. ([@davewongillies](https://github.com/davewongillies))

## 0.7.5 (2019-04-03)

### Enhancements

- Update RDS DB size list ([#269](https://github.com/terraform-linters/tflint/pull/269))
- Add M5 and R5 families to ElastiCache ([#270](https://github.com/terraform-linters/tflint/pull/270))

### Others

- Add go report card ([#261](https://github.com/terraform-linters/tflint/pull/261))
- automate the installation of tflint on linux ([#267](https://github.com/terraform-linters/tflint/pull/267))

## 0.7.4 (2019-02-09)

### Enhancements

- Add support for db.m5 series db types ([#258](https://github.com/terraform-linters/tflint/pull/258))

## 0.7.3 (2018-12-28)

### Enhancements

- Update ec2-instances-info dependency ([#257](https://github.com/terraform-linters/tflint/pull/257))

### Others

- Add "features" word to docs for people explicitly looking ([#237](https://github.com/terraform-linters/tflint/pull/237))

## 0.7.2 (2018-08-26)

### Enhancements

- Update valid instance list ([#226](https://github.com/terraform-linters/tflint/pull/226))

## 0.7.1 (2018-07-19)

### Bugfix

- Add missing db instances as valid types ([#214](https://github.com/terraform-linters/tflint/pull/214))
- Update valid instance types ([#215](https://github.com/terraform-linters/tflint/pull/215))

### Others

- Migrate to dep from Glide ([#208](https://github.com/terraform-linters/tflint/pull/208))
- Add `rule` section in README ([#213](https://github.com/terraform-linters/tflint/pull/213))

## 0.7.0 (2018-06-04)

### Enhancements

- Add new `rule` configuration syntax ([#197](https://github.com/terraform-linters/tflint/pull/197))

### Others

- Recommend `rule` syntax instead of `ignore_rules` in README ([#200](https://github.com/terraform-linters/tflint/pull/200))

## 0.6.0 (2018-05-18)

### Enhancements

- Support terraform.workspace variable ([#181](https://github.com/terraform-linters/tflint/pull/181))
- Accept glob and multiple input ([#183](https://github.com/terraform-linters/tflint/pull/183))
- Fallback to config under the home directory ([#186](https://github.com/terraform-linters/tflint/pull/186))
- Add new --quiet option ([#190](https://github.com/terraform-linters/tflint/pull/190))

### Changes

- Remove aws_instance_not_specified_iam_profile ([#180](https://github.com/terraform-linters/tflint/pull/180))

### Bugfix

- Handle color for Windows ([#184](https://github.com/terraform-linters/tflint/pull/184))
- Fix interpolation checking ([#189](https://github.com/terraform-linters/tflint/pull/189))
- Detect pinned sources using regular expressions ([#194](https://github.com/terraform-linters/tflint/pull/194))

### Others

- AppVeyor :rocket: ([#185](https://github.com/terraform-linters/tflint/pull/185))
- Add note for installation ([#196](https://github.com/terraform-linters/tflint/pull/196))

## 0.5.4 (2018-01-07)

### Bugfix

- Handle empty config file ([#166](https://github.com/terraform-linters/tflint/pull/166))

## 0.5.3 (2017-12-09)

### Enhancements

- Support module path for v0.11.0 ([#161](https://github.com/terraform-linters/tflint/pull/161))
- Ignore module initialization when settings `ignore_module` ([#163](https://github.com/terraform-linters/tflint/pull/163))

## 0.5.2 (2017-11-12)

### Enhancements

- Use `cristim/ec2-instances-info` instead of hard-coded list ([#159](https://github.com/terraform-linters/tflint/pull/159))

### BugFix

- Use `strings.Trim` instead of `strings.Replace` ([#158](https://github.com/terraform-linters/tflint/pull/158))

### Others

- Set Docker container default workdir to /data ([#152](https://github.com/terraform-linters/tflint/pull/152))
- Add ca-certificates to Docker image for TLS requests to AWS ([#155](https://github.com/terraform-linters/tflint/pull/155))

## 0.5.1 (2017-10-18)

Re-release due to [#151](https://github.com/terraform-linters/tflint/issues/151)  
There is no change in the code from v0.5.0

## 0.5.0 (2017-10-14)

Minor version update. This release includes environment variable support.

### Enhancements

- Support variables from environment variables ([#147](https://github.com/terraform-linters/tflint/pull/147))
- Support moudle path for v0.10.7 ([#149](https://github.com/terraform-linters/tflint/pull/149))

### Others

- Add Makefile target for creating docker image ([#145](https://github.com/terraform-linters/tflint/pull/145))
- Update Go version ([#146](https://github.com/terraform-linters/tflint/pull/146))

## 0.4.3 (2017-09-30)

Patch version update. This release includes Terraform v0.10.6 supports.

### Enhancements

- Add G3 instances support ([#139](https://github.com/terraform-linters/tflint/pull/139))
- Support new digest module path ([#144](https://github.com/terraform-linters/tflint/pull/144))

### Others

- Fix unclear error messages ([#137](https://github.com/terraform-linters/tflint/pull/137))

## 0.4.2 (2017-08-03)

Patch version update. This release includes a hotfix.

### BugFix

- Fix panic for integer variables interpolation ([#131](https://github.com/terraform-linters/tflint/pull/131))

## 0.4.1 (2017-07-29)

Patch version update. This release includes terraform meta information interpolation syntax support.

### NewDetectors

- Add AwsECSClusterDuplicateNameDetector ([#128](https://github.com/terraform-linters/tflint/pull/128))

### Enhancements

- Support "${terraform.env}" syntax ([#126](https://github.com/terraform-linters/tflint/pull/126))
- Environment state handling ([#127](https://github.com/terraform-linters/tflint/pull/127))

### Others

- Update deps ([#130](https://github.com/terraform-linters/tflint/pull/130))

## 0.4.0 (2017-07-09)

Minor version update. This release includes big core API changes.

### Enhancements

- Overrides module ([#118](https://github.com/terraform-linters/tflint/pull/118))
- Add document link and detector name on output ([#122](https://github.com/terraform-linters/tflint/pull/122))
- Add Terraform version options ([#123](https://github.com/terraform-linters/tflint/pull/123))
- Report `aws_instance_not_specified_iam_profile` only when `terraform_version` is less than 0.8.8 ([#124](https://github.com/terraform-linters/tflint/pull/124))

### Others

- Provide abstract HCL access ([#112](https://github.com/terraform-linters/tflint/pull/112))
- Fix override logic ([#117](https://github.com/terraform-linters/tflint/pull/117))
- Fix some output messages and documentation ([#125](https://github.com/terraform-linters/tflint/pull/125))

## 0.3.6 (2017-06-05)

Patch version update. This release includes hotfix for module evaluation.

### BugFix

- DO NOT USE Evaluator :bow: ([#114](https://github.com/terraform-linters/tflint/pull/114))

### Others

- Add HCL syntax highlighting in README ([#110](https://github.com/terraform-linters/tflint/pull/110))
- Update README.md ([#111](https://github.com/terraform-linters/tflint/pull/111))

## 0.3.5 (2017-04-23)

Patch version update. This release includes new detectors and bugfix for module.

### NewDetectors

- Module source pinned ref check ([#100](https://github.com/terraform-linters/tflint/pull/100))
- Add AwsCloudWatchMetricAlarmInvalidUnitDetector ([#108](https://github.com/terraform-linters/tflint/pull/108))

### Enhancements

- Support F1 instances ([#107](https://github.com/terraform-linters/tflint/pull/107))

### BugFix

- Interpolate module attributes ([#105](https://github.com/terraform-linters/tflint/pull/105))

### Others

- Improve CLI ([#102](https://github.com/terraform-linters/tflint/pull/102))
- Add integration test ([#106](https://github.com/terraform-linters/tflint/pull/106))

## 0.3.4 (2017-04-10)

Patch version update. This release includes new detectors for `aws_route`

### NewDetectors

- Add AwsRouteInvalidRouteTableDetector ([#90](https://github.com/terraform-linters/tflint/pull/90))
- Add AwsRouteNotSpecifiedTargetDetector ([#91](https://github.com/terraform-linters/tflint/pull/91))
- Add AwsRouteSpecifiedMultipleTargetsDetector ([#92](https://github.com/terraform-linters/tflint/pull/92))
- Add AwsRouteInvalidGatewayDetector ([#93](https://github.com/terraform-linters/tflint/pull/93))
- Add AwsRouteInvalidEgressOnlyGatewayDetector ([#94](https://github.com/terraform-linters/tflint/pull/94))
- Add AwsRouteInvalidNatGatewayDetector ([#95](https://github.com/terraform-linters/tflint/pull/95))
- Add AwsRouteInvalidVpcPeeringConnectionDetector ([#96](https://github.com/terraform-linters/tflint/pull/96))
- Add AwsRouteInvalidInstanceDetector ([#97](https://github.com/terraform-linters/tflint/pull/97))
- Add AwsRouteInvalidNetworkInterfaceDetector ([#98](https://github.com/terraform-linters/tflint/pull/98))

### BugFix

- Fix panic when security groups are on EC2-Classic ([#89](https://github.com/terraform-linters/tflint/pull/89))

### Others

- Transfer from hakamadare/tflint to terraform-linters/tflint ([#84](https://github.com/terraform-linters/tflint/pull/84))

## 0.3.3 (2017-04-02)

Patch version update. This release includes support for shared credentials.

### Enhancements

- Support shared credentials ([#79](https://github.com/terraform-linters/tflint/pull/79))
- Add checkstyle format ([#82](https://github.com/terraform-linters/tflint/pull/82))

### Others

- Add NOTE to aws_instance_not_specified_iam_profile ([#81](https://github.com/terraform-linters/tflint/pull/81))
- Refactoring for default printer ([#83](https://github.com/terraform-linters/tflint/pull/83))

## 0.3.2 (2017-03-25)

Patch version update. This release includes hotfix.

### BugFix

- Fix panic when parsing empty list ([#78](https://github.com/terraform-linters/tflint/pull/78))

### Others

- Fix unstable test ([#74](https://github.com/terraform-linters/tflint/pull/74))
- Update README to reference Homebrew tap ([#75](https://github.com/terraform-linters/tflint/pull/75))

## 0.3.1 (2017-03-12)

Patch version update. This release includes support for tfvars.

### Enhancements

- Support I3 instance types ([#66](https://github.com/terraform-linters/tflint/pull/66))
- Support TFVars ([#67](https://github.com/terraform-linters/tflint/pull/67))

### Others

- Add Dockerfile ([#59](https://github.com/terraform-linters/tflint/pull/59))
- Fix link ([#60](https://github.com/terraform-linters/tflint/pull/60))
- Update help message ([#61](https://github.com/terraform-linters/tflint/pull/61))
- Move cache from detector to awsclient ([#62](https://github.com/terraform-linters/tflint/pull/62))
- Refactoring detector ([#65](https://github.com/terraform-linters/tflint/pull/65))
- glide up ([#68](https://github.com/terraform-linters/tflint/pull/68))
- Update go version ([#69](https://github.com/terraform-linters/tflint/pull/69))

## 0.3.0 (2017-02-12)

Minor version update. This release includes core enhancements for terraform state file.

### NewDetectors

- Add RDS readable password detector ([#46](https://github.com/terraform-linters/tflint/pull/46))
- Add duplicate security group name detector ([#49](https://github.com/terraform-linters/tflint/pull/49))
- Add duplicate ALB name detector ([#52](https://github.com/terraform-linters/tflint/pull/52))
- Add duplicate ELB name detector ([#54](https://github.com/terraform-linters/tflint/pull/54))
- Add duplicate DB Instance Identifier Detector ([#55](https://github.com/terraform-linters/tflint/pull/55))
- Add duplicate ElastiCache Cluster ID detector ([#56](https://github.com/terraform-linters/tflint/pull/56))

### Enhancements

- Interpret TFState ([#48](https://github.com/terraform-linters/tflint/pull/48))
- Add --fast option ([#58](https://github.com/terraform-linters/tflint/pull/58))

### BugFix

- r4.xlarge is valid type ([#43](https://github.com/terraform-linters/tflint/pull/43))

### Others

- Add sideci.yml ([#42](https://github.com/terraform-linters/tflint/pull/42))
- Update README ([#50](https://github.com/terraform-linters/tflint/pull/50))
- SideCI Settings ([#57](https://github.com/terraform-linters/tflint/pull/57))

## 0.2.1 (2017-01-10)

Patch version update. This release includes new argument options.

### NewDetectors

- add db instance invalid type detector ([#32](https://github.com/terraform-linters/tflint/pull/32))
- add rds previous type detector ([#33](https://github.com/terraform-linters/tflint/pull/33))
- add invalid type detector for elasticache ([#34](https://github.com/terraform-linters/tflint/pull/34))
- add previous type detector for elasticache ([#35](https://github.com/terraform-linters/tflint/pull/35))

### Enhancements

- Return error code when issue exists ([#31](https://github.com/terraform-linters/tflint/pull/31))

### Others

- fix install version ([#30](https://github.com/terraform-linters/tflint/pull/30))
- CLI Test By Interface ([#36](https://github.com/terraform-linters/tflint/pull/36))
- Fix --error-with-issues description ([#37](https://github.com/terraform-linters/tflint/pull/37))
- glide up ([#38](https://github.com/terraform-linters/tflint/pull/38))

## 0.2.0 (2016-12-24)

Minor version update. This release includes enhancements and several fixes

### New Detectors

- add AWS Instance Invalid AMI deep detector ([#7](https://github.com/terraform-linters/tflint/pull/7))
- add invalid key name deep detector ([#11](https://github.com/terraform-linters/tflint/pull/11))
- add invalid subnet deep detector ([#12](https://github.com/terraform-linters/tflint/pull/12))
- add invalid vpc security group deep detector ([#13](https://github.com/terraform-linters/tflint/pull/13))
- add invalid security group detector for ELB ([#16](https://github.com/terraform-linters/tflint/pull/16))
- add invalid subnet detector for ELB ([#17](https://github.com/terraform-linters/tflint/pull/17))
- add invalid instance detector for ELB ([#18](https://github.com/terraform-linters/tflint/pull/18))
- add invalid security group detector for ALB ([#20](https://github.com/terraform-linters/tflint/pull/20))
- add invalid subnet detector for ALB ([#21](https://github.com/terraform-linters/tflint/pull/21))
- add invalid security group detector for RDS ([#22](https://github.com/terraform-linters/tflint/pull/22))
- add invalid DB subnet group detector for RDS ([#23](https://github.com/terraform-linters/tflint/pull/23))
- add invalid parameter group detector for RDS ([#24](https://github.com/terraform-linters/tflint/pull/24))
- add invalid option group detector for RDS ([#25](https://github.com/terraform-linters/tflint/pull/25))
- add invalid parameter group detector for ElastiCache ([#27](https://github.com/terraform-linters/tflint/pull/27))
- add invalid subnet group detector for ElastiCache ([#28](https://github.com/terraform-linters/tflint/pull/28))
- add invalid security group detector for ElastiCache ([#29](https://github.com/terraform-linters/tflint/pull/29))

### Enhancements

- Support t2 and r4 types ([#5](https://github.com/terraform-linters/tflint/pull/5))
- Improve ineffecient module detector method ([#10](https://github.com/terraform-linters/tflint/pull/10))
- do not call API when target resources are not found ([#15](https://github.com/terraform-linters/tflint/pull/15))
- support list type variables evaluation ([#19](https://github.com/terraform-linters/tflint/pull/19))

### Bug Fixes

- Fix panic deep detecting with module ([#8](https://github.com/terraform-linters/tflint/pull/8))

### Others

- Fix `Fatalf` format in test ([#3](https://github.com/terraform-linters/tflint/pull/3))
- Remove Zero width space in README.md ([#4](https://github.com/terraform-linters/tflint/pull/4))
- Fix typos ([#6](https://github.com/terraform-linters/tflint/pull/6))
- documentation ([#26](https://github.com/terraform-linters/tflint/pull/26))

## 0.1.0 (2016-11-27)

Initial release

### Added

- Add Fundamental features

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Nothing
