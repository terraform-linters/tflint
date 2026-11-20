package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindWorkingDirs(t *testing.T) {
	// The tree used by every case:
	//
	//   root/
	//     a/
	//     b/
	//       .git/    (hidden, must be skipped)
	//       c/
	//     .hidden/
	//       d/
	root := t.TempDir()
	for _, dir := range []string{"a", "b/.git", "b/c", ".hidden/d"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		wd        string
		opts      Options
		want      []string
		wantError bool
	}{
		{
			name: "recursive from the current directory",
			wd:   ".",
			opts: Options{Recursive: true},
			want: []string{".", "a", "b", "b/c"},
		},
		{
			name: "recursive with --chdir",
			wd:   ".",
			opts: Options{Recursive: true, Chdir: "b"},
			want: []string{"b", "b/c"},
		},
		{
			name: "recursive with --chdir pointing at the parent directory",
			wd:   "a",
			opts: Options{Recursive: true, Chdir: ".."},
			want: []string{"..", "../a", "../b", "../b/c"},
		},
		{
			name: "recursive with --chdir pointing at the current directory with a trailing separator",
			wd:   ".",
			opts: Options{Recursive: true, Chdir: "./"},
			want: []string{"./", "a", "b", "b/c"},
		},
		{
			name: "recursive with --chdir pointing at a hidden directory",
			wd:   ".",
			opts: Options{Recursive: true, Chdir: ".hidden"},
			want: []string{".hidden", ".hidden/d"},
		},
		{
			name: "not recursive",
			wd:   ".",
			opts: Options{Chdir: ".."},
			want: []string{".."},
		},
		{
			name:      "recursive with a missing --chdir",
			wd:        ".",
			opts:      Options{Recursive: true, Chdir: "missing"},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Chdir(filepath.Join(root, test.wd))

			got, err := findWorkingDirs(test.opts)
			if test.wantError {
				if err == nil {
					t.Fatalf("expected an error, but got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			want := make([]string, len(test.want))
			for i, dir := range test.want {
				want[i] = filepath.FromSlash(dir)
			}
			// filepath.WalkDir passes the root through unchanged, but cleans the paths of everything below it.
			if test.opts.Chdir == "./" {
				want[0] = "./"
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("working directories mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
