import = "aws-sdk-go/models/apis/iam/2010-05-08/api-2.json"

mapping "aws_iam_access_key" {
  user    = existingUserNameType
  pgp_key = stringType
  status  = statusType
}

mapping "aws_iam_account_alias" {
  account_alias = any // accountAliasType
}

mapping "aws_iam_account_password_policy" {
  allow_users_to_change_password = booleanType
  hard_expiry                    = booleanObjectType
  max_password_age               = maxPasswordAgeType
  minimum_password_length        = minimumPasswordLengthType
  password_reuse_prevention      = passwordReusePreventionType
  require_lowercase_characters   = booleanType
  require_numbers                = booleanType
  require_symbols                = booleanType
  require_uppercase_characters   = booleanType
}

mapping "aws_iam_group" {
  name = groupNameType
  path = pathType
}

mapping "aws_iam_group_membership" {
  name  = any
  users = any
  group = groupNameType
}

mapping "aws_iam_group_policy" {
  policy      = policyDocumentType
  name        = policyNameType
  name_prefix = any
  group       = groupNameType
}

mapping "aws_iam_group_policy_attachment" {
  group      = groupNameType
  policy_arn = arnType
}

mapping "aws_iam_instance_profile" {
  name        = instanceProfileNameType
  name_prefix = any
  path        = pathType
  role        = roleNameType
}

mapping "aws_iam_openid_connect_provider" {
  url             = OpenIDConnectProviderUrlType
  client_id_list  = clientIDListType
  thumbprint_list = thumbprintListType
}

mapping "aws_iam_policy" {
  description = policyDescriptionType
  name        = policyNameType
  name_prefix = any
  path        = policyPathType
  policy      = policyDocumentType
}

mapping "aws_iam_policy_attachment" {
  name       = any
  users      = any
  roles      = any
  groups     = any
  policy_arn = arnType
}

mapping "aws_iam_role" {
  name                  = roleNameType
  name_prefix           = any
  assume_role_policy    = policyDocumentType
  force_detach_policies = any
  path                  = pathType
  description           = roleDescriptionType
  max_session_duration  = roleMaxSessionDurationType
  permissions_boundary  = arnType
  tags                  = tagListType
}

mapping "aws_iam_role_policy" {
  name        = policyNameType
  name_prefix = any
  policy      = policyDocumentType
  role        = roleNameType
}

mapping "aws_iam_role_policy_attachment" {
  role       = roleNameType
  policy_arn = arnType
}

mapping "aws_iam_saml_provider" {
  name                   = SAMLProviderNameType
  saml_metadata_document = SAMLMetadataDocumentType
}

mapping "aws_iam_server_certificate" {
  name              = serverCertificateNameType
  name_prefix       = any
  certificate_body  = certificateBodyType
  certificate_chain = certificateChainType
  private_key       = privateKeyType
  path              = pathType
}

mapping "aws_iam_service_linked_role" {
  aws_service_name = groupNameType
  custom_suffix    = customSuffixType
  description      = roleDescriptionType
}

mapping "aws_iam_user" {
  name                 = userNameType
  path                 = pathType
  permissions_boundary = arnType
  force_destroy        = any
  tags                 = tagListType
}

mapping "aws_iam_user_group_membership" {
  user   = userNameType
  groups = any
}

mapping "aws_iam_user_login_profile" {
  user                    = userNameType
  pgp_key                 = any
  password_length         = any
  password_reset_required = booleanType
}

mapping "aws_iam_user_policy" {
  policy      = policyDocumentType
  name        = policyNameType
  name_prefix = any
  user        = existingUserNameType
}

mapping "aws_iam_user_policy_attachment" {
  user       = existingUserNameType
  policy_arn = arnType
}

mapping "aws_iam_user_ssh_key" {
  username   = userNameType
  encoding   = encodingType
  public_key = publicKeyMaterialType
  status     = statusType
}
