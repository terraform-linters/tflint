import = "aws-sdk-go/models/apis/acm-pca/2017-08-22/api-2.json"

mapping "aws_acmpca_certificate_authority" {
  type = CertificateAuthorityType
}

test "aws_acmpca_certificate_authority" "type" {
  ok = "SUBORDINATE"
  ng = "ORDINATE"
}
