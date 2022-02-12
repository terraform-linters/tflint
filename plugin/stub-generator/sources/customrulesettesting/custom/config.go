package custom

import "github.com/hashicorp/hcl/v2"

type Config struct {
	// From .tflint.hcl
	DeepCheck bool  `hcl:"deep_check,optional"`
	Auth      *Auth `hcl:"auth,block"`

	// From provider config
	Zone       string
	Annotation string

	Remain hcl.Body `hcl:",remain"`
}

type Auth struct {
	Token string `hcl:"token"`
}
