package workingdir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestReadsResolveWithoutConfiningPaths(t *testing.T) {
	root := t.TempDir()
	base := afero.NewOsFs()
	for _, name := range []string{"module/file", "sibling/file"} {
		path := filepath.Join(root, name)
		if err := base.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := afero.WriteFile(base, path, []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
	}
	fs := New(base, filepath.Join(root, "module"))
	for path, want := range map[string]string{
		"file":                              "module/file",
		"../sibling/file":                   "sibling/file",
		filepath.Join(root, "sibling/file"): "sibling/file",
	} {
		got, err := afero.ReadFile(fs, path)
		if err != nil || string(got) != want {
			t.Errorf("read %s: %q, %v", path, got, err)
		}
		f, err := fs.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		if f.Name() != path {
			t.Errorf("filename changed: %s", f.Name())
		}
		f.Close()
		if _, err := fs.OpenFile(path, os.O_RDWR, 0644); !os.IsPermission(err) {
			t.Errorf("write should be rejected: %v", err)
		}
	}
}
