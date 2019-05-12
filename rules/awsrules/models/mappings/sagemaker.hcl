import = "aws-sdk-go/models/apis/sagemaker/2017-07-24/api-2.json"

mapping "aws_sagemaker_endpoint" {
  endpoint_config_name = EndpointConfigName
  name                 = EndpointName
  tags                 = TagList
}

mapping "aws_sagemaker_endpoint_configuration" {
  production_variants = ProductionVariantList
  kms_key_arn         = KmsKeyId
  name                = EndpointConfigName
  tags                = TagList
}

mapping "aws_sagemaker_model" {
  name                     = ModelName
  primary_container        = ContainerDefinition
  execution_role_arn       = RoleArn
  container                = ContainerDefinitionList
  enable_network_isolation = Boolean
  vpc_config               = VpcConfig
  tags                     = TagList
}

mapping "aws_sagemaker_notebook_instance" {
  name                  = NotebookInstanceName
  role_arn              = RoleArn
  instance_type         = InstanceType
  subnet_id             = SubnetId
  security_groups       = SecurityGroupIds 
  kms_key_id            = KmsKeyId
  lifecycle_config_name = NotebookInstanceLifecycleConfigName
  tags                  = TagList
}

mapping "aws_sagemaker_notebook_instance_lifecycle_configuration" {
  name      = NotebookInstanceLifecycleConfigName
  on_create = NotebookInstanceLifecycleConfigList
  on_start  = NotebookInstanceLifecycleConfigList
}
