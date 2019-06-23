import = "aws-sdk-go/models/apis/inspector/2016-02-16/api-2.json"

mapping "aws_inspector_assessment_target" {
  name               = AssessmentTargetName
  resource_group_arn = Arn
}

mapping "aws_inspector_assessment_template" {
  name               = AssessmentTemplateName
  target_arn         = Arn
  duration           = AssessmentRunDuration
  rules_package_arns = AssessmentTemplateRulesPackageArnList
}

mapping "aws_inspector_resource_group" {
  tags = ResourceGroupTags
}
