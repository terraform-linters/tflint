import = "aws-sdk-go/models/apis/cloudfront/2019-03-26/api-2.json"

mapping "aws_cloudfront_distribution" {
  http_version = HttpVersion
  price_class  = PriceClass
}

test "aws_cloudfront_distribution" "http_version" {
  ok = "http2"
  ng = "http1.2"
}

test "aws_cloudfront_distribution" "price_class" {
  ok = "PriceClass_All"
  ng = "PriceClass_300"
}
