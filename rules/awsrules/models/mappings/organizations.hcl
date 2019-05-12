import = "aws-sdk-go/models/apis/organizations/2016-11-28/api-2.json"

mapping "aws_organizations_account" {
  name                       = AccountName
  email                      = Email
  iam_user_access_to_billing = IAMUserAccessToBilling
  parent_id                  = ParentId
  role_name                  = RoleName
}

mapping "aws_organizations_organization" {
  aws_service_access_principals = any
  enabled_policy_types          = PolicyTypes
  feature_set                   = OrganizationFeatureSet
}

mapping "aws_organizations_organizational_unit" {
  name      = OrganizationalUnitName
  parent_id = ParentId
}

mapping "aws_organizations_policy" {
  content     = PolicyContent
  name        = PolicyName
  description = PolicyDescription
  type        = PolicyType
}

mapping "aws_organizations_policy_attachment" {
  policy_id = PolicyId
  target_id = PolicyTargetId
}
