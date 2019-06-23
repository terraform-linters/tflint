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
