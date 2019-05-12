import = "aws-sdk-go/models/apis/elastictranscoder/2012-09-25/api-2.json"

mapping "aws_elastictranscoder_pipeline" {
  aws_kms_key_arn              = KeyArn
  content_config               = PipelineOutputConfig
  content_config_permissions   = Permissions
  input_bucket                 = BucketName
  name                         = Name
  notifications                = Notifications
  output_bucket                = BucketName
  role                         = Role
  thumbnail_config             = PipelineOutputConfig
  thumbnail_config_permissions = Permissions
}

mapping "aws_elastictranscoder_preset" {
  audio               = AudioParameters
  audio_codec_options = AudioCodecOptions
  container           = PresetContainer
  description         = Description
  name                = Name
  thumbnails          = Thumbnails
  video               = VideoParameters
  video_watermarks    = PresetWatermarks
  video_codec_options = any
}

test "aws_elastictranscoder_preset" "container" {
  ok = "mp4"
  ng = "mp1"
}
