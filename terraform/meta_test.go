package terraform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDataDir(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "default",
			want: ".terraform",
		},
		{
			name: "TF_DATA_DIR",
			env:  map[string]string{"TF_DATA_DIR": ".tfdata"},
			want: ".tfdata",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.env {
				t.Setenv(k, v)
			}

			got := DataDir()
			if test.want != got {
				t.Errorf("want: %s, got: %s", test.want, got)
			}
		})
	}
}

func TestModuleDir(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "default",
			want: filepath.Join(".terraform", "modules"),
		},
		{
			name: "TF_DATA_DIR",
			env:  map[string]string{"TF_DATA_DIR": ".tfdata"},
			want: filepath.Join(".tfdata", "modules"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.env {
				t.Setenv(k, v)
			}

			got := ModuleDir()
			if test.want != got {
				t.Errorf("want: %s, got: %s", test.want, got)
			}
		})
	}
}

func TestModuleManifestPath(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "default",
			want: filepath.Join(".terraform", "modules", "modules.json"),
		},
		{
			name: "TF_DATA_DIR",
			env:  map[string]string{"TF_DATA_DIR": ".tfdata"},
			want: filepath.Join(".tfdata", "modules", "modules.json"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.env {
				t.Setenv(k, v)
			}

			got := ModuleManifestPath()
			if test.want != got {
				t.Errorf("want: %s, got: %s", test.want, got)
			}
		})
	}
}

func TestWorkspace(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		dir  string
		env  map[string]string
		want string
	}{
		{
			name: "default",
			want: "default",
		},
		{
			name: "TF_WORKSPACE",
			env:  map[string]string{"TF_WORKSPACE": "dev"},
			want: "dev",
		},
		{
			name: "env file",
			dir:  filepath.Join(currentDir, "test-fixtures", "workspace"),
			want: "staging",
		},
		{
			name: "TF_DATA_DIR",
			dir:  filepath.Join(currentDir, "test-fixtures", "workspace"),
			env:  map[string]string{"TF_DATA_DIR": ".terraform_production"},
			want: "production",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.dir != "" {
				if err := os.Chdir(test.dir); err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := os.Chdir(currentDir); err != nil {
						t.Fatal(err)
					}
				}()
			}

			for k, v := range test.env {
				t.Setenv(k, v)
			}

			got := Workspace()
			if test.want != got {
				t.Errorf("want: %s, got: %s", test.want, got)
			}
		})
	}
}
