import = "aws-sdk-go/models/apis/config/2014-11-12/api-2.json"

mapping "aws_config_aggregate_authorization" {
  account_id = AccountId
  region     = AwsRegion
}

mapping "aws_config_configuration_aggregator" {
  name = ConfigurationAggregatorName
}

mapping "aws_config_config_rule" {
  name                        = ConfigRuleName
  description                 = EmptiableStringWithCharLimit256
  input_parameters            = StringWithCharLimit1024
  maximum_execution_frequency = MaximumExecutionFrequency
}

mapping "aws_config_configuration_recorder" {
  name = RecorderName
}

mapping "aws_config_configuration_recorder_status" {
  name = RecorderName
}

mapping "aws_config_organization_managed_rule" {
  name                        = StringWithCharLimit64
  rule_identifier             = StringWithCharLimit256
  description                 = StringWithCharLimit256Min0
  excluded_accounts           = ExcludedAccounts
  input_parameters            = StringWithCharLimit2048
  maximum_execution_frequency = MaximumExecutionFrequency
  resource_id_scope           = StringWithCharLimit768
  resource_types_scope        = ResourceTypesScope
  tag_key_scope               = StringWithCharLimit128
  tag_value_scope             = StringWithCharLimit256
}

mapping "aws_config_organization_custom_rule" {
  lambda_function_arn          = StringWithCharLimit256
  name                         = StringWithCharLimit64
  trigger_types                = OrganizationConfigRuleTriggerTypes
  description                  = StringWithCharLimit256Min0
  excluded_accounts            = ExcludedAccounts
  input_parameters             = StringWithCharLimit2048
  maximum_execution_frequency  = MaximumExecutionFrequency
  resource_id_scope            = StringWithCharLimit768
  resource_types_scope         = ResourceTypesScope
  tag_key_scope                = StringWithCharLimit128
  tag_value_scope              = StringWithCharLimit256
}

mapping "aws_config_delivery_channel" {
  name = ChannelName
}

test "aws_config_aggregate_authorization" "account_id" {
  ok = "012345678910"
  ng = "01234567891"
}

test "aws_config_configuration_aggregator" "name" {
  ok = "example"
  ng = "example.com"
}

test "aws_config_config_rule" "maximum_execution_frequency" {
  ok = "One_Hour"
  ng = "Hour"
}
