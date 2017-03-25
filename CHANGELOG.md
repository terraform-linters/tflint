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
