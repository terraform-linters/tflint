package config

import "strings"

type Config struct {
	Debug        bool
	IgnoreModule map[string]bool
	IgnoreRule   map[string]bool
}

func Init(ignoreModule string, ignoreRule string) *Config {
	var ignoreModules []string = strings.Split(ignoreModule, ",")
	var ignoreRules []string = strings.Split(ignoreRule, ",")
	ignoreModuleMap := map[string]bool{}
	ignoreRuleMap := map[string]bool{}
	for _, m := range ignoreModules {
		ignoreModuleMap[m] = true
	}
	for _, r := range ignoreRules {
		ignoreRuleMap["Detect"+r] = true
	}

	return &Config{
		Debug:        false,
		IgnoreModule: ignoreModuleMap,
		IgnoreRule:   ignoreRuleMap,
	}
}
