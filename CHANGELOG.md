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
