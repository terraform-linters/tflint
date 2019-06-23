import = "aws-sdk-go/models/apis/mediastore/2017-09-01/api-2.json"

mapping "aws_media_store_container" {
  name = ContainerName
}

mapping "aws_media_store_container_policy" {
  container_name = ContainerName
  policy         = any
}
