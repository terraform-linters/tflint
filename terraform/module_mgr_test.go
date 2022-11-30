package terraform

import (
	"path/filepath"
	"testing"
)

func Test_moduleManifestPath(t *testing.T) {
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

			got := moduleManifestPath()
			if test.want != got {
				t.Errorf("want: %s, got: %s", test.want, got)
			}
		})
	}
}
