import = "aws-sdk-go/models/apis/cognito-idp/2016-04-18/api-2.json"

mapping "aws_cognito_identity_provider" {
  user_pool_id  = UserPoolIdType
  provider_name = ProviderNameType
  provider_type = IdentityProviderTypeType
}

mapping "aws_cognito_resource_server" {
  identifier = ResourceServerIdentifierType
  name       = ResourceServerNameType
}

mapping "aws_cognito_user_group" {
  name         = GroupNameType
  user_pool_id = UserPoolIdType
  description  = DescriptionType
  precedence   = PrecedenceType
  role_arn     = ArnType
}

mapping "aws_cognito_user_pool" {
  alias_attributes           = AliasAttributesListType
  auto_verified_attributes   = VerifiedAttributesListType
  name                       = UserPoolNameType
  email_verification_subject = EmailVerificationSubjectType
  email_verification_message = EmailVerificationMessageType
  mfa_configuration          = UserPoolMfaType
  sms_authentication_message = SmsVerificationMessageType
  sms_verification_message   = SmsVerificationMessageType
}

mapping "aws_cognito_user_pool_client" {
  default_redirect_uri   = RedirectUrlType
  name                   = ClientNameType
  refresh_token_validity = RefreshTokenValidityType
  user_pool_id           = UserPoolIdType
}

mapping "aws_cognito_user_pool_domain" {
  domain          = DomainType
  user_pool_id    = UserPoolIdType
  certificate_arn = ArnType
}

test "aws_cognito_identity_provider" "user_pool_id" {
  ok = "foo_bar"
  ng = "foobar"
}

test "aws_cognito_identity_provider" "provider_name" {
  ok = "Google"
  ng = "\t"
}

test "aws_cognito_identity_provider" "provider_type" {
  ok = "LoginWithAmazon"
  ng = "Apple"
}

test "aws_cognito_resource_server" "identifier" {
  ok = "https://example.com"
  ng = "\t"
}

test "aws_cognito_resource_server" "name" {
  ok = "example"
  ng = "example/server"
}

test "aws_cognito_user_group" "name" {
  ok = "user-group"
  ng = "user\tgroup"
}

test "aws_cognito_user_group" "role_arn" {
  ok = "arn:aws:iam::123456789012:role/s3access"
  ng = "aws:iam::123456789012:instance-profile/s3access-profile"
}

test "aws_cognito_user_pool" "name" {
  ok = "mypool"
  ng = "my/pool"
}

test "aws_cognito_user_pool" "email_verification_message" {
  ok = "Verification code is {####}"
  ng = "Verification code"
}

test "aws_cognito_user_pool" "mfa_configuration" {
  ok = "ON"
  ng = "IN"
}

test "aws_cognito_user_pool" "sms_authentication_message" {
  ok = "Authentication code is {####}"
  ng = "Authentication code"
}

test "aws_cognito_user_pool" "sms_verification_message" {
  ok = "Verification code is {####}"
  ng = "Verification code"
}

test "aws_cognito_user_pool_client" "default_redirect_uri" {
  ok = "https://example.com/callback"
  ng = "https://example com"
}

test "aws_cognito_user_pool_client" "name" {
  ok = "client"
  ng = "client/example"
}

test "aws_cognito_user_pool_domain" "domain" {
  ok = "auth"
  ng = "auth example"
}
