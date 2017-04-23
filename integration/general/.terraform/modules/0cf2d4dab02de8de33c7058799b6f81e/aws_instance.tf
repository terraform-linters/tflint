variable "instance_types" {
  type = "map"
}

resource "aws_instance" "invalid" {
  ami           = "ami-22ce4934"                     // aws_instance_not_specified_iam_profile
  instance_type = "${var.instance_types["invalid"]}" // aws_instance_invalid_type

  root_block_device = {
    volume_size = "16" // aws_instance_default_standard_volume
  }
}

resource "aws_instance" "previous" {
  ami                  = "ami-22ce4934"
  iam_instance_profile = "test-profile"
  instance_type        = "${var.instance_types["previous"]}" // aws_instance_previous_type

  root_block_device = {
    volume_type = "default"
    volume_size = "16"
  }
}
