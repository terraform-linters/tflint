plugin "testing" {
  enabled = true
}

plugin "terraform" {
  enabled = false
}

config {
  varfile = ["from_config.tfvars"]
}
