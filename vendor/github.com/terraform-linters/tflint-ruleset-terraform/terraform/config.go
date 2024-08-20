package terraform

// Config is the configuration for the ruleset.
type Config struct {
	Preset string `hclext:"preset,optional"`
}
