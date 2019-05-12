import = "aws-sdk-go/models/apis/apigateway/2015-07-09/api-2.json"

mapping "aws_api_gateway_gateway_response" {
  status_code = StatusCode
}

mapping "aws_api_gateway_integration_response" {
  status_code = StatusCode
  content_handling = ContentHandlingStrategy
}

mapping "aws_api_gateway_method_response" {
  status_code = StatusCode
}

mapping "aws_api_gateway_authorizer" {
  type = AuthorizerType
}

mapping "aws_api_gateway_gateway_response" {
  response_type = GatewayResponseType
}

mapping "aws_api_gateway_integration" {
  type = IntegrationType
}

mapping "aws_api_gateway_integration" {
  connection_type  = ConnectionType
  content_handling = ContentHandlingStrategy
}

mapping "aws_api_gateway_rest_api" {
  api_key_source = ApiKeySourceType
}

mapping "aws_api_gateway_stage" {
  cache_cluster_size = CacheClusterSize
}

test "aws_api_gateway_gateway_response" "status_code" {
  ok = "200"
  ng = "004"
}

test "aws_api_gateway_authorizer" "type" {
  ok = "TOKEN"
  ng = "RESPONSE"
}

test "aws_api_gateway_gateway_response" "response_type" {
  ok = "UNAUTHORIZED"
  ng = "4XX"
}

test "aws_api_gateway_integration" "type" {
  ok = "HTTP"
  ng = "AWS_HTTP"
}

test "aws_api_gateway_integration" "connection_type" {
  ok = "INTERNET"
  ng = "INTRANET"
}

test "aws_api_gateway_integration" "content_handling" {
  ok = "CONVERT_TO_BINARY"
  ng = "CONVERT_TO_FILE"
}

test "aws_api_gateway_rest_api" "api_key_source" {
  ok = "AUTHORIZER"
  ng = "BODY"
}

test "aws_api_gateway_stage" "cache_cluster_size" {
  ok = "6.1"
  ng = "6.2"
}
