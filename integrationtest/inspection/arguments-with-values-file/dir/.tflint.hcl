plugin "testing" {
  enabled = true
}

config {
  // relative path is resolved based on the current directory, not the config file path.
  varfile = ["current_dir.tfvars"]
}
