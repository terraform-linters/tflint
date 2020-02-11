import = "aws-sdk-go/models/apis/acm/2015-12-08/api-2.json"

mapping "aws_acm_certificate" {
  // domain_name            = DomainNameString
  subject_alternative_names = DomainList
  // validation_method      = ValidationMethod
  private_key               = PrivateKey
  certificate_body          = CertificateBody
  certificate_chain         = CertificateChain
  tags                      = TagList
}
