import = "aws-sdk-go/models/apis/transfer/2018-11-05/api-2.json"

mapping "aws_transfer_server" {
  endpoint_details       = EndpointDetails
  endpoint_type          = EndpointType
  invocation_role        = Role
  url                    = Url
  identity_provider_type = IdentityProviderType
  logging_role           = Role
  force_destroy          = any
  tags                   = Tags
}

mapping "aws_transfer_ssh_key" {
  server_id = ServerId
  user_name = UserName
  body      = SshPublicKeyBody
}

mapping "aws_transfer_user" {
  server_id      = ServerId
  user_name      = UserName
  home_directory = HomeDirectory
  policy         = Policy
  role           = Role
  tags           = Tags
}
