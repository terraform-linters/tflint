package tflint

// Config is a TFLint configuration applied to the plugin.
type Config struct {
	Rules             map[string]*RuleConfig
	DisabledByDefault bool
	Only              []string
	Fix               bool
}

// RuleConfig is a TFLint's rule configuration.
type RuleConfig struct {
	Name    string
	Enabled bool
}
