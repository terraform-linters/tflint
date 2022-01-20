variable "ssl_certificates" {
  type = "map"
  default = {
    lorem-elb-us-west-1 = "lorem"
    ipsum-elb-us-east-1 = "ipsum"
    dolor-elb-us-east-4 = "dolor"
  }
}

