package plugin

import (
	"github.com/wata727/tflint/tflint"
)

// Plugin is a mechanism for adding third-party rules.
// Each plugin must be built with `--buildmode=plugin`.
// @see https://golang.org/pkg/plugin/
//
// Each plugin must have a top-level function named `Name`, `Version` and `NewRules`.
// The return value of NewRules must be a structure that satisfies the Rule interface.
type Plugin struct {
	Name    string
	Version string
	Rules   []Rule
}

// Rule is an interface that each plugin should implement.
type Rule interface {
	Name() string
	Enabled() bool
	Severity() string
	Link() string
	Check(runner *tflint.Runner) error
}

// NewRules returns all available plugin rules.
func NewRules() ([]Rule, error) {
	plugins, err := Find()
	if err != nil {
		return nil, err
	}

	rules := []Rule{}
	for _, p := range plugins {
		rules = append(rules, p.Rules...)
	}
	return rules, nil
}
