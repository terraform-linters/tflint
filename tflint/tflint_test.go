package tflint

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func withinFixtureDir(t *testing.T, dir string, test func()) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Chdir(filepath.Join(currentDir, "test-fixtures", dir))
	test()
}

func testRunnerWithOsFs(t *testing.T, config *Config) *Runner {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	loader, err := terraform.NewLoader(afero.Afero{Fs: afero.NewOsFs()}, originalWd)
	if err != nil {
		t.Fatal(err)
	}

	cfg, diags := loader.LoadConfig(".", config.CallModuleType)
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	runner, err := NewRunner(originalWd, config, map[string]Annotations{}, cfg, map[string]*terraform.InputValue{})
	if err != nil {
		t.Fatal(err)
	}

	return runner
}

func testRunnerWithAnnotations(t *testing.T, files map[string]string, annotations map[string]Annotations) *Runner {
	config := EmptyConfig()
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	for name, src := range files {
		err := fs.WriteFile(name, []byte(src), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	loader, err := terraform.NewLoader(fs, originalWd)
	if err != nil {
		t.Fatal(err)
	}

	cfg, diags := loader.LoadConfig(".", config.CallModuleType)
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	runner, err := NewRunner(originalWd, config, annotations, cfg, map[string]*terraform.InputValue{})
	if err != nil {
		t.Fatal(err)
	}

	return runner
}

func moduleConfig() *Config {
	c := EmptyConfig()
	c.CallModuleType = terraform.CallAllModule
	return c
}
