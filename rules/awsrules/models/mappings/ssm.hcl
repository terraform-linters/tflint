import = "aws-sdk-go/models/apis/ssm/2014-11-06/api-2.json"

mapping "aws_ssm_activation" {
  name               = DefaultInstanceName
  description        = ActivationDescription
  expiration_date    = ExpirationDate
  iam_role           = IamRole
  registration_limit = RegistrationLimit
  tags               = TagList
}

mapping "aws_ssm_association" {
  name                = DocumentARN
  association_name    = AssociationName
  document_version    = DocumentVersion
  instance_id         = InstanceId
  output_location     = InstanceAssociationOutputLocation
  parameters          = Parameters
  schedule_expression = ScheduleExpression
  targets             = Targets
  compliance_severity = AssociationComplianceSeverity
  max_concurrency     = MaxConcurrency
  max_errors          = MaxErrors
}

mapping "aws_ssm_document" {
  name            = DocumentName
  content         = DocumentContent
  document_format = DocumentFormat
  document_type   = DocumentType
  permissions     = any
  tags            = TagList
}

mapping "aws_ssm_maintenance_window" {
  name                       = MaintenanceWindowName
  schedule                   = MaintenanceWindowSchedule
  cutoff                     = MaintenanceWindowCutoff
  duration                   = MaintenanceWindowDurationHours
  allow_unassociated_targets = MaintenanceWindowAllowUnassociatedTargets
  enabled                    = MaintenanceWindowEnabled
  end_date                   = MaintenanceWindowStringDateTime
  schedule_timezone          = MaintenanceWindowTimezone
  start_date                 = MaintenanceWindowStringDateTime
  tags                       = TagList
}

mapping "aws_ssm_maintenance_window_target" {
  window_id         = MaintenanceWindowId
  name              = MaintenanceWindowName
  description       = MaintenanceWindowDescription
  resource_type     = MaintenanceWindowResourceType
  targets           = Targets
  owner_information = OwnerInformation
}

mapping "aws_ssm_maintenance_window_task" {
  window_id        = MaintenanceWindowId
  max_concurrency  = MaxConcurrency
  max_errors       = MaxErrors
  task_type        = MaintenanceWindowTaskType
  task_arn         = MaintenanceWindowTaskArn
  service_role_arn = ServiceRole
  name             = MaintenanceWindowName
  description      = MaintenanceWindowDescription
  targets          = Targets
  priority         = MaintenanceWindowTaskPriority
}

mapping "aws_ssm_patch_baseline" {
  name                              = BaselineName
  description                       = BaselineDescription
  operating_system                  = OperatingSystem
  approved_patches_compliance_level = PatchComplianceLevel
  approved_patches                  = PatchIdList
  rejected_patches                  = PatchIdList
  global_filter                     = PatchFilterGroup
  approval_rule                     = PatchRuleGroup
}

mapping "aws_ssm_patch_group" {
  baseline_id = BaselineId
  patch_group = PatchGroup
}

mapping "aws_ssm_parameter" {
  name            = PSParameterName
  type            = ParameterType
  value           = PSParameterValue
  description     = ParameterDescription
  tier            = ParameterTier
  key_id          = ParameterKeyId
  overwrite       = Boolean
  allowed_pattern = AllowedPattern
  tags            = TagList
}

mapping "aws_ssm_resource_data_sync" {
  name           = ResourceDataSyncName
  s3_destination = ResourceDataSyncS3Destination
}
