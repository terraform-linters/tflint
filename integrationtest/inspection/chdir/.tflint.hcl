plugin "testing" {
  enabled = true
}

plugin "terraform" {
  enabled = false
}

config {
  varfile = ["dir/from_config.tfvars"]
}
