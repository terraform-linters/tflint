plugin "customrulesettesting" {
  enabled = true
  deep_check = true

  auth {
    token = "SECRET_TOKEN"
  }
}

plugin "aws" {
  enabled = false
}
