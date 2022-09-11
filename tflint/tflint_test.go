package tflint

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
	defer func() {
		if err = os.Chdir(currentDir); err != nil {
			t.Fatal(err)
		}
	}()

	if err = os.Chdir(filepath.Join(currentDir, "test-fixtures", dir)); err != nil {
		t.Fatal(err)
	}

	test()
}

func testRunnerWithOsFs(t *testing.T, config *Config) *Runner {
	loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, config)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := loader.LoadConfig(".")
	if err != nil {
		t.Fatal(err)
	}

	runner, err := NewRunner(config, map[string]Annotations{}, cfg, map[string]*terraform.InputValue{})
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

	loader, err := NewLoader(fs, config)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := loader.LoadConfig(".")
	if err != nil {
		t.Fatal(err)
	}

	runner, err := NewRunner(config, annotations, cfg, map[string]*terraform.InputValue{})
	if err != nil {
		t.Fatal(err)
	}

	return runner
}

func moduleConfig() *Config {
	c := EmptyConfig()
	c.Module = true
	return c
}

func newLine() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}
