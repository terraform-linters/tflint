package custom

type Config struct {
	// From .tflint.hcl
	DeepCheck bool  `hclext:"deep_check,optional"`
	Auth      *Auth `hclext:"auth,block"`

	// From provider config
	Zone       string
	Annotation string
}

type Auth struct {
	Token string `hclext:"token"`
}
