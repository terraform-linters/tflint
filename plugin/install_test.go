package plugin

import (
	"os"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_Install(t *testing.T) {
	original := PluginRoot
	PluginRoot = t.TempDir()
	defer func() { PluginRoot = original }()

	config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
		Name:        "aws",
		Enabled:     true,
		Version:     "0.4.0",
		Source:      "github.com/terraform-linters/tflint-ruleset-aws",
		SourceOwner: "terraform-linters",
		SourceRepo:  "tflint-ruleset-aws",
	})

	path, err := config.Install()
	if err != nil {
		t.Fatalf("Failed to install: %s", err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open installed binary: %s", err)
	}
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat installed binary: %s", err)
	}
	file.Close()

	expected := "tflint-ruleset-aws" + fileExt()
	if info.Name() != expected {
		t.Fatalf("Installed binary name is invalid: expected=%s, got=%s", expected, info.Name())
	}
}
