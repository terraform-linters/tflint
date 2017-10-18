## 0.5.1 (2017-10-18)

Re-release due to [#151](https://github.com/wata727/tflint/issues/151)  
There is no change in the code from v0.5.0

## 0.5.0 (2017-10-14)

Minor version update. This release includes environment variable support.

### Enhancements

- Support variables from environment variables ([#147](https://github.com/wata727/tflint/pull/147))
- Support moudle path for v0.10.7 ([#149](https://github.com/wata727/tflint/pull/149))

## Others

- Add Makefile target for creating docker image ([#145](https://github.com/wata727/tflint/pull/145))
- Update Go version ([#146](https://github.com/wata727/tflint/pull/146))

## 0.4.3 (2017-09-30)

Patch version update. This release includes Terraform v0.10.6 supports.

### Enhancements

- Add G3 instances support ([#139](https://github.com/wata727/tflint/pull/139))
- Support new digest module path ([#144](https://github.com/wata727/tflint/pull/144))

### Others

- Fix unclear error messages ([#137](https://github.com/wata727/tflint/pull/137))

## 0.4.2 (2017-08-03)

Patch version update. This release includes a hotfix.

### BugFix

- Fix panic for integer variables interpolation ([#131](https://github.com/wata727/tflint/pull/131))

## 0.4.1 (2017-07-29)

Patch version update. This release includes terraform meta information interpolation syntax support.

### NewDetectors

- Add AwsECSClusterDuplicateNameDetector ([#128](https://github.com/wata727/tflint/pull/128))

### Enhancements

- Support "${terraform.env}" syntax ([#126](https://github.com/wata727/tflint/pull/126))
- Environment state handling ([#127](https://github.com/wata727/tflint/pull/127))

### Others

- Update deps ([#130](https://github.com/wata727/tflint/pull/130))

## 0.4.0 (2017-07-09)

Minor version update. This release includes big core API changes.

### Enhancements

- Overrides module ([#118](https://github.com/wata727/tflint/pull/118))
- Add document link and detector name on output ([#122](https://github.com/wata727/tflint/pull/122))
- Add Terraform version options ([#123](https://github.com/wata727/tflint/pull/123))
- Report `aws_instance_not_specified_iam_profile` only when `terraform_version` is less than 0.8.8 ([#124](https://github.com/wata727/tflint/pull/124))

### Others

- Provide abstract HCL access ([#112](https://github.com/wata727/tflint/pull/112))
- Fix override logic ([#117](https://github.com/wata727/tflint/pull/117))
- Fix some output messages and documentation ([#125](https://github.com/wata727/tflint/pull/125))

## 0.3.6 (2017-06-05)

Patch version update. This release includes hotfix for module evaluation.

### BugFix

- DO NOT USE Evaluator :bow: ([#114](https://github.com/wata727/tflint/pull/114))

### Others

- Add HCL syntax highlighting in README ([#110](https://github.com/wata727/tflint/pull/110))
- Update README.md ([#111](https://github.com/wata727/tflint/pull/111))

## 0.3.5 (2017-04-23)

Patch version update. This release includes new detectors and bugfix for module.

### NewDetectors

- Module source pinned ref check ([#100](https://github.com/wata727/tflint/pull/100))
- Add AwsCloudWatchMetricAlarmInvalidUnitDetector ([#108](https://github.com/wata727/tflint/pull/108))

### Enhancements

- Support F1 instances ([#107](https://github.com/wata727/tflint/pull/107))

### BugFix

- Interpolate module attributes ([#105](https://github.com/wata727/tflint/pull/105))

### Others

- Improve CLI ([#102](https://github.com/wata727/tflint/pull/102))
- Add integration test ([#106](https://github.com/wata727/tflint/pull/106))

## 0.3.4 (2017-04-10)

Patch version update. This release includes new detectors for `aws_route`

### NewDetectors

- Add AwsRouteInvalidRouteTableDetector ([#90](https://github.com/wata727/tflint/pull/90))
- Add AwsRouteNotSpecifiedTargetDetector ([#91](https://github.com/wata727/tflint/pull/91))
- Add AwsRouteSpecifiedMultipleTargetsDetector ([#92](https://github.com/wata727/tflint/pull/92))
- Add AwsRouteInvalidGatewayDetector ([#93](https://github.com/wata727/tflint/pull/93))
- Add AwsRouteInvalidEgressOnlyGatewayDetector ([#94](https://github.com/wata727/tflint/pull/94))
- Add AwsRouteInvalidNatGatewayDetector ([#95](https://github.com/wata727/tflint/pull/95))
- Add AwsRouteInvalidVpcPeeringConnectionDetector ([#96](https://github.com/wata727/tflint/pull/96))
- Add AwsRouteInvalidInstanceDetector ([#97](https://github.com/wata727/tflint/pull/97))
- Add AwsRouteInvalidNetworkInterfaceDetector ([#98](https://github.com/wata727/tflint/pull/98))

### BugFix

- Fix panic when security groups are on EC2-Classic ([#89](https://github.com/wata727/tflint/pull/89))

### Others

- Transfer from hakamadare/tflint to wata727/tflint ([#84](https://github.com/wata727/tflint/pull/84))

## 0.3.3 (2017-04-02)

Patch version update. This release includes support for shared credentials.

### Enhancements

- Support shared credentials ([#79](https://github.com/wata727/tflint/pull/79))
- Add checkstyle format ([#82](https://github.com/wata727/tflint/pull/82))

### Others

- Add NOTE to aws_instance_not_specified_iam_profile ([#81](https://github.com/wata727/tflint/pull/81))
- Refactoring for default printer ([#83](https://github.com/wata727/tflint/pull/83))

## 0.3.2 (2017-03-25)

Patch version update. This release includes hotfix.

### BugFix

- Fix panic when parsing empty list ([#78](https://github.com/wata727/tflint/pull/78))

### Others

- Fix unstable test ([#74](https://github.com/wata727/tflint/pull/74))
- Update README to reference Homebrew tap ([#75](https://github.com/wata727/tflint/pull/75))

## 0.3.1 (2017-03-12)

Patch version update. This release includes support for tfvars.

### Enhancements

- Support I3 instance types ([#66](https://github.com/wata727/tflint/pull/66))
- Support TFVars ([#67](https://github.com/wata727/tflint/pull/67))

### Others

- Add Dockerfile ([#59](https://github.com/wata727/tflint/pull/59))
- Fix link ([#60](https://github.com/wata727/tflint/pull/60))
- Update help message ([#61](https://github.com/wata727/tflint/pull/61))
- Move cache from detector to awsclient ([#62](https://github.com/wata727/tflint/pull/62))
- Refactoring detector ([#65](https://github.com/wata727/tflint/pull/65))
- glide up ([#68](https://github.com/wata727/tflint/pull/68))
- Update go version ([#69](https://github.com/wata727/tflint/pull/69))

## 0.3.0 (2017-02-12)

Minor version update. This release includes core enhancements for terraform state file.

### NewDetectors

- Add RDS readable password detector ([#46](https://github.com/wata727/tflint/pull/46))
- Add duplicate security group name detector ([#49](https://github.com/wata727/tflint/pull/49))
- Add duplicate ALB name detector ([#52](https://github.com/wata727/tflint/pull/52))
- Add duplicate ELB name detector ([#54](https://github.com/wata727/tflint/pull/54))
- Add duplicate DB Instance Identifier Detector ([#55](https://github.com/wata727/tflint/pull/55))
- Add duplicate ElastiCache Cluster ID detector ([#56](https://github.com/wata727/tflint/pull/56))

### Enhancements

- Interpret TFState ([#48](https://github.com/wata727/tflint/pull/48))
- Add --fast option ([#58](https://github.com/wata727/tflint/pull/58))

### BugFix

- r4.xlarge is valid type ([#43](https://github.com/wata727/tflint/pull/43))

### Others

- Add sideci.yml ([#42](https://github.com/wata727/tflint/pull/42))
- Update README ([#50](https://github.com/wata727/tflint/pull/50))
- SideCI Settings ([#57](https://github.com/wata727/tflint/pull/57))

## 0.2.1 (2017-01-10)

Patch version update. This release includes new argument options.

### NewDetectors

- add db instance invalid type detector ([#32](https://github.com/wata727/tflint/pull/32))
- add rds previous type detector ([#33](https://github.com/wata727/tflint/pull/33))
- add invalid type detector for elasticache ([#34](https://github.com/wata727/tflint/pull/34))
- add previous type detector for elasticache ([#35](https://github.com/wata727/tflint/pull/35))

### Enhancements

- Return error code when issue exists ([#31](https://github.com/wata727/tflint/pull/31))

### Others

- fix install version ([#30](https://github.com/wata727/tflint/pull/30))
- CLI Test By Interface ([#36](https://github.com/wata727/tflint/pull/36))
- Fix --error-with-issues description ([#37](https://github.com/wata727/tflint/pull/37))
- glide up ([#38](https://github.com/wata727/tflint/pull/38))

## 0.2.0 (2016-12-24)

Minor version update. This release includes enhancements and several fixes

### New Detectors

- add AWS Instance Invalid AMI deep detector ([#7](https://github.com/wata727/tflint/pull/7))
- add invalid key name deep detector ([#11](https://github.com/wata727/tflint/pull/11))
- add invalid subnet deep detector ([#12](https://github.com/wata727/tflint/pull/12))
- add invalid vpc security group deep detector ([#13](https://github.com/wata727/tflint/pull/13))
- add invalid security group detector for ELB ([#16](https://github.com/wata727/tflint/pull/16))
- add invalid subnet detector for ELB ([#17](https://github.com/wata727/tflint/pull/17))
- add invalid instance detector for ELB ([#18](https://github.com/wata727/tflint/pull/18))
- add invalid security group detector for ALB ([#20](https://github.com/wata727/tflint/pull/20))
- add invalid subnet detector for ALB ([#21](https://github.com/wata727/tflint/pull/21))
- add invalid security group detector for RDS ([#22](https://github.com/wata727/tflint/pull/22))
- add invalid DB subnet group detector for RDS ([#23](https://github.com/wata727/tflint/pull/23))
- add invalid parameter group detector for RDS ([#24](https://github.com/wata727/tflint/pull/24))
- add invalid option group detector for RDS ([#25](https://github.com/wata727/tflint/pull/25))
- add invalid parameter group detector for ElastiCache ([#27](https://github.com/wata727/tflint/pull/27))
- add invalid subnet group detector for ElastiCache ([#28](https://github.com/wata727/tflint/pull/28))
- add invalid security group detector for ElastiCache ([#29](https://github.com/wata727/tflint/pull/29))

### Enhancements

- Support t2 and r4 types ([#5](https://github.com/wata727/tflint/pull/5))
- Improve ineffecient module detector method ([#10](https://github.com/wata727/tflint/pull/10))
- do not call API when target resources are not found ([#15](https://github.com/wata727/tflint/pull/15))
- support list type variables evaluation ([#19](https://github.com/wata727/tflint/pull/19))

### Bug Fixes

- Fix panic deep detecting with module ([#8](https://github.com/wata727/tflint/pull/8))

### Others

- Fix `Fatalf` format in test ([#3](https://github.com/wata727/tflint/pull/3))
- Remove Zero width space in README.md ([#4](https://github.com/wata727/tflint/pull/4))
- Fix typos ([#6](https://github.com/wata727/tflint/pull/6))
- documentation ([#26](https://github.com/wata727/tflint/pull/26))

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
