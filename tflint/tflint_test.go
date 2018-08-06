package tflint

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func loadConfigHelper(dir string) (*configs.Config, error) {
	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		return nil, err
	}

	mod, diags := loader.Parser().LoadConfigDir(dir)
	if diags.HasErrors() {
		return nil, diags
	}
	cfg, diags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
	if diags.HasErrors() {
		return nil, diags
	}

	return cfg, nil
}

func extractAttributeHelper(key string, cfg *configs.Config) (*hcl.Attribute, error) {
	resource := cfg.Module.ManagedResources["null_resource.test"]
	if resource == nil {
		return nil, errors.New("Expected resource is not found")
	}
	body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: key,
			},
		},
	})
	if diags.HasErrors() {
		return nil, diags
	}
	attribute := body.Attributes[key]
	if attribute == nil {
		return nil, fmt.Errorf("Expected attribute is not found: %s", key)
	}
	return attribute, nil
}
