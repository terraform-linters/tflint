import = "aws-sdk-go/models/apis/servicecatalog/2015-12-10/api-2.json"

mapping "aws_servicecatalog_portfolio" {
  name          = PortfolioDisplayName
  description   = PortfolioDescription
  provider_name = ProviderName
  tags          = AddTags
}
