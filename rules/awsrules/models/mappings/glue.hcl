import = "aws-sdk-go/models/apis/glue/2017-03-31/api-2.json"

mapping "aws_glue_catalog_database" {
  name         = any // NameString
  catalog_id   = any // CatalogIdString
  description  = any // DescriptionString
  location_uri = any // URI
  parameters   = ParametersMap
}

mapping "aws_glue_catalog_table" {
  name               = any // NameString
  database_name      = any // NameString
  catalog_id         = any // CatalogIdString
  description        = any // DescriptionString
  owner              = any // NameString
  retention          = NonNegativeInteger
  storage_descriptor = StorageDescriptor
  partition_keys     = ColumnList
  view_original_text = ViewTextString
  view_expanded_text = ViewTextString
  table_type         = TableTypeString
  parameters         = ParametersMap
}

mapping "aws_glue_classifier" {
  grok_classifier = CreateGrokClassifierRequest
  json_classifier = CreateJsonClassifierRequest
  name            = any // NameString
  xml_classifier  = CreateXMLClassifierRequest
}

mapping "aws_glue_connection" {
  catalog_id                       = any // CatalogIdString
  connection_properties            = ConnectionProperties
  connection_type                  = ConnectionType
  description                      = any // DescriptionString
  match_criteria                   = MatchCriteria
  name                             = any // NameString
  physical_connection_requirements = PhysicalConnectionRequirements
}

mapping "aws_glue_crawler" {
  database_name          = DatabaseName
  name                   = any // NameString
  role                   = Role
  classifiers            = ClassifierNameList
  configuration          = CrawlerConfiguration
  description            = any // DescriptionString
  dynamodb_target        = DynamoDBTargetList
  jdbc_target            = JdbcTargetList
  s3_target              = S3TargetList
  schedule               = CronExpression
  schema_change_policy   = SchemaChangePolicy
  table_prefix           = TablePrefix
  security_configuration = CrawlerSecurityConfiguration
}

mapping "aws_glue_job" {
  allocated_capacity     = IntegerValue
  command                = JobCommand
  connections            = ConnectionsList
  default_arguments      = GenericMap
  description            = any // DescriptionString
  execution_property     = ExecutionProperty
  max_capacity           = NullableDouble
  max_retries            = MaxRetries
  name                   = any // NameString
  role_arn               = RoleString
  timeout                = Timeout
  security_configuration = any // NameString
}

mapping "aws_glue_security_configuration" {
  encryption_configuration = EncryptionConfiguration
  name                     = any // NameString
}

mapping "aws_glue_trigger" {
  actions     = ActionList
  description = any // DescriptionString
  enabled     = Boolean
  name        = any // NameString
  predicate   = Predicate
  schedule    = GenericString
  type        = TriggerType
}
