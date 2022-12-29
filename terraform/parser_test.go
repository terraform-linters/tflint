package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcltest"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

func TestLoadConfigDir(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		baseDir string
		dir     string
		want    *Module
	}{
		{
			name: "HCL native files",
			files: map[string]string{
				"main.tf":          "",
				"main_override.tf": "",
				"override.tf":      "",
			},
			baseDir: ".",
			dir:     ".",
			want: &Module{
				SourceDir: ".",
				primaries: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: "main.tf"}})},
				},
				overrides: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: "main_override.tf"}})},
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: "override.tf"}})},
				},
				Sources: map[string][]byte{
					"main.tf":          {},
					"main_override.tf": {},
					"override.tf":      {},
				},
				Files: map[string]*hcl.File{
					"main.tf":          {Body: hcl.EmptyBody()},
					"main_override.tf": {Body: hcl.EmptyBody()},
					"override.tf":      {Body: hcl.EmptyBody()},
				},
			},
		},
		{
			name: "HCL JSON files",
			files: map[string]string{
				"main.tf.json":          "{}",
				"main_override.tf.json": "{}",
				"override.tf.json":      "{}",
			},
			baseDir: ".",
			dir:     ".",
			want: &Module{
				SourceDir: ".",
				primaries: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: "main.tf.json"}})},
				},
				overrides: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: "main_override.tf.json"}})},
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: "override.tf.json"}})},
				},
				Sources: map[string][]byte{
					"main.tf.json":          []byte("{}"),
					"main_override.tf.json": []byte("{}"),
					"override.tf.json":      []byte("{}"),
				},
				Files: map[string]*hcl.File{
					"main.tf.json":          {Body: hcl.EmptyBody()},
					"main_override.tf.json": {Body: hcl.EmptyBody()},
					"override.tf.json":      {Body: hcl.EmptyBody()},
				},
			},
		},
		{
			name: "with base dir",
			files: map[string]string{
				"main.tf":          "",
				"main_override.tf": "",
				"override.tf":      "",
			},
			baseDir: "foo",
			dir:     ".",
			want: &Module{
				SourceDir: ".",
				primaries: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: filepath.Join("foo", "main.tf")}})},
				},
				overrides: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: filepath.Join("foo", "main_override.tf")}})},
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: filepath.Join("foo", "override.tf")}})},
				},
				Sources: map[string][]byte{
					filepath.Join("foo", "main.tf"):          {},
					filepath.Join("foo", "main_override.tf"): {},
					filepath.Join("foo", "override.tf"):      {},
				},
				Files: map[string]*hcl.File{
					filepath.Join("foo", "main.tf"):          {Body: hcl.EmptyBody()},
					filepath.Join("foo", "main_override.tf"): {Body: hcl.EmptyBody()},
					filepath.Join("foo", "override.tf"):      {Body: hcl.EmptyBody()},
				},
			},
		},
		{
			name: "with dir",
			files: map[string]string{
				filepath.Join("bar", "main.tf"):          "",
				filepath.Join("bar", "main_override.tf"): "",
				filepath.Join("bar", "override.tf"):      "",
			},
			baseDir: ".",
			dir:     "bar",
			want: &Module{
				SourceDir: "bar",
				primaries: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: filepath.Join("bar", "main.tf")}})},
				},
				overrides: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: filepath.Join("bar", "main_override.tf")}})},
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: filepath.Join("bar", "override.tf")}})},
				},
				Sources: map[string][]byte{
					filepath.Join("bar", "main.tf"):          {},
					filepath.Join("bar", "main_override.tf"): {},
					filepath.Join("bar", "override.tf"):      {},
				},
				Files: map[string]*hcl.File{
					filepath.Join("bar", "main.tf"):          {Body: hcl.EmptyBody()},
					filepath.Join("bar", "main_override.tf"): {Body: hcl.EmptyBody()},
					filepath.Join("bar", "override.tf"):      {Body: hcl.EmptyBody()},
				},
			},
		},
		{
			name: "with basedir + dir",
			files: map[string]string{
				filepath.Join("bar", "main.tf"):          "",
				filepath.Join("bar", "main_override.tf"): "",
				filepath.Join("bar", "override.tf"):      "",
			},
			baseDir: "foo",
			dir:     "bar",
			want: &Module{
				SourceDir: "bar",
				primaries: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: filepath.Join("foo", "bar", "main.tf")}})},
				},
				overrides: []*hcl.File{
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: filepath.Join("foo", "bar", "main_override.tf")}})},
					{Body: hcltest.MockBody(&hcl.BodyContent{MissingItemRange: hcl.Range{Filename: filepath.Join("foo", "bar", "override.tf")}})},
				},
				Sources: map[string][]byte{
					filepath.Join("foo", "bar", "main.tf"):          {},
					filepath.Join("foo", "bar", "main_override.tf"): {},
					filepath.Join("foo", "bar", "override.tf"):      {},
				},
				Files: map[string]*hcl.File{
					filepath.Join("foo", "bar", "main.tf"):          {Body: hcl.EmptyBody()},
					filepath.Join("foo", "bar", "main_override.tf"): {Body: hcl.EmptyBody()},
					filepath.Join("foo", "bar", "override.tf"):      {Body: hcl.EmptyBody()},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			for name, content := range test.files {
				if err := fs.WriteFile(name, []byte(content), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}
			parser := NewParser(fs)

			mod, diags := parser.LoadConfigDir(test.baseDir, test.dir)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			if mod.SourceDir != test.want.SourceDir {
				t.Errorf("SourceDir: want=%s, got=%s", test.want.SourceDir, mod.SourceDir)
			}

			opt := cmpopts.SortSlices(func(x, y string) bool { return x > y })

			primaries := make([]string, len(mod.primaries))
			for i, f := range mod.primaries {
				primaries[i] = f.Body.MissingItemRange().Filename
			}
			primariesWant := make([]string, len(test.want.primaries))
			for i, f := range test.want.primaries {
				primariesWant[i] = f.Body.MissingItemRange().Filename
			}
			if diff := cmp.Diff(primaries, primariesWant, opt); diff != "" {
				t.Errorf(diff)
			}

			overrides := make([]string, len(mod.overrides))
			for i, f := range mod.overrides {
				overrides[i] = f.Body.MissingItemRange().Filename
			}
			overridesWant := make([]string, len(test.want.overrides))
			for i, f := range test.want.overrides {
				overridesWant[i] = f.Body.MissingItemRange().Filename
			}
			if diff := cmp.Diff(overrides, overridesWant, opt); diff != "" {
				t.Errorf(diff)
			}

			if diff := cmp.Diff(mod.Sources, test.want.Sources); diff != "" {
				t.Errorf(diff)
			}

			files := []string{}
			for name := range mod.Files {
				files = append(files, name)
			}
			filesWant := []string{}
			for name := range test.want.Files {
				filesWant = append(filesWant, name)
			}
			if diff := cmp.Diff(files, filesWant, opt); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestLoadConfigDirFiles(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		baseDir string
		dir     string
		want    []string
	}{
		{
			name: "HCL native files",
			files: map[string]string{
				"main.tf":          "",
				"main_override.tf": "",
				"override.tf":      "",
			},
			baseDir: ".",
			dir:     ".",
			want: []string{
				"main.tf",
				"main_override.tf",
				"override.tf",
			},
		},
		{
			name: "HCL JSON files",
			files: map[string]string{
				"main.tf.json":          "{}",
				"main_override.tf.json": "{}",
				"override.tf.json":      "{}",
			},
			baseDir: ".",
			dir:     ".",
			want: []string{
				"main.tf.json",
				"main_override.tf.json",
				"override.tf.json",
			},
		},
		{
			name: "with base dir",
			files: map[string]string{
				"main.tf":          "",
				"main_override.tf": "",
				"override.tf":      "",
			},
			baseDir: "foo",
			dir:     ".",
			want: []string{
				filepath.Join("foo", "main.tf"),
				filepath.Join("foo", "main_override.tf"),
				filepath.Join("foo", "override.tf"),
			},
		},
		{
			name: "with dir",
			files: map[string]string{
				filepath.Join("bar", "main.tf"):          "",
				filepath.Join("bar", "main_override.tf"): "",
				filepath.Join("bar", "override.tf"):      "",
			},
			baseDir: ".",
			dir:     "bar",
			want: []string{
				filepath.Join("bar", "main.tf"),
				filepath.Join("bar", "main_override.tf"),
				filepath.Join("bar", "override.tf"),
			},
		},
		{
			name: "with basedir + dir",
			files: map[string]string{
				filepath.Join("bar", "main.tf"):          "",
				filepath.Join("bar", "main_override.tf"): "",
				filepath.Join("bar", "override.tf"):      "",
			},
			baseDir: "foo",
			dir:     "bar",
			want: []string{
				filepath.Join("foo", "bar", "main.tf"),
				filepath.Join("foo", "bar", "main_override.tf"),
				filepath.Join("foo", "bar", "override.tf"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			for name, content := range test.files {
				if err := fs.WriteFile(name, []byte(content), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}
			parser := NewParser(fs)

			files, diags := parser.LoadConfigDirFiles(test.baseDir, test.dir)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opt := cmpopts.SortSlices(func(x, y string) bool { return x > y })

			got := []string{}
			for name := range files {
				got = append(got, name)
			}
			if diff := cmp.Diff(got, test.want, opt); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestLoadValuesFile(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		baseDir string
		path    string
		want    map[string]cty.Value
		sources map[string][]byte
	}{
		{
			name: "default",
			files: map[string]string{
				"terraform.tfvars": `foo="bar"`,
			},
			baseDir: ".",
			path:    "terraform.tfvars",
			want: map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			},
			sources: map[string][]byte{
				"terraform.tfvars": []byte(`foo="bar"`),
			},
		},
		{
			name: "with base dir",
			files: map[string]string{
				"terraform.tfvars": `foo="bar"`,
			},
			baseDir: "baz",
			path:    "terraform.tfvars",
			want: map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			},
			sources: map[string][]byte{
				filepath.Join("baz", "terraform.tfvars"): []byte(`foo="bar"`),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			for name, content := range test.files {
				if err := fs.WriteFile(name, []byte(content), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}
			parser := NewParser(fs)

			got, diags := parser.LoadValuesFile(test.baseDir, test.path)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opt := cmpopts.IgnoreUnexported(cty.Value{})
			if diff := cmp.Diff(got, test.want, opt); diff != "" {
				t.Error(diff)
			}
			if diff := cmp.Diff(parser.Sources(), test.sources); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestIsConfigDir(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		baseDir string
		dir     string
		want    bool
	}{
		{
			name: "HCL native files (primary)",
			files: map[string]string{
				"main.tf": "",
			},
			baseDir: ".",
			dir:     ".",
			want:    true,
		},
		{
			name: "HCL native files (override)",
			files: map[string]string{
				"override.tf": "",
			},
			baseDir: ".",
			dir:     ".",
			want:    true,
		},
		{
			name: "HCL JSON files (primary)",
			files: map[string]string{
				"main.tf.json": "{}",
			},
			baseDir: ".",
			dir:     ".",
			want:    true,
		},
		{
			name: "HCL JSON files (override)",
			files: map[string]string{
				"override.tf.json": "{}",
			},
			baseDir: ".",
			dir:     ".",
			want:    true,
		},
		{
			name: "non-HCL files",
			files: map[string]string{
				"README.md": "",
			},
			baseDir: ".",
			dir:     ".",
			want:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			for name, content := range test.files {
				if err := fs.WriteFile(name, []byte(content), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}
			parser := NewParser(fs)

			got := parser.IsConfigDir(test.baseDir, test.dir)

			if got != test.want {
				t.Errorf("want=%t, got=%t", test.want, got)
			}
		})
	}
}
