package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

// Parser is a fork of configs.Parser. This is the main interface to read
// configuration files and other related files from disk.
//
// It retains a cache of all files that are loaded so that they can be used
// to create source code snippets in diagnostics, etc.
type Parser struct {
	fs afero.Afero
	p  *hclparse.Parser
}

// NewParser creates and returns a new Parser that reads files from the given
// filesystem. If a nil filesystem is passed then the system's "real" filesystem
// will be used, via afero.OsFs.
func NewParser(fs afero.Fs) *Parser {
	if fs == nil {
		fs = afero.OsFs{}
	}

	return &Parser{
		fs: afero.Afero{Fs: fs},
		p:  hclparse.NewParser(),
	}
}

// LoadConfigDir reads the .tf and .tf.json files in the given directory and
// then combines these files into a single Module.
//
// If this method returns nil, that indicates that the given directory does not
// exist at all or could not be opened for some reason. Callers may wish to
// detect this case and ignore the returned diagnostics so that they can
// produce a more context-aware error message in that case.
//
// If this method returns a non-nil module while error diagnostics are returned
// then the module may be incomplete but can be used carefully for static
// analysis.
//
// This file does not consider a directory with no files to be an error, and
// will simply return an empty module in that case.
//
// .tf files are parsed using the HCL native syntax while .tf.json files are
// parsed using the HCL JSON syntax.
//
// If a baseDir is passed, the loaded files are assumed to be loaded from that
// directory.
func (p *Parser) LoadConfigDir(baseDir, dir string) (*Module, hcl.Diagnostics) {
	primaries, overrides, diags := p.configDirFiles(baseDir, dir)
	if diags.HasErrors() {
		return nil, diags
	}

	mod := NewEmptyModule()
	mod.primaries = make([]*hcl.File, len(primaries))
	mod.overrides = make([]*hcl.File, len(overrides))

	for i, path := range primaries {
		f, loadDiags := p.loadHCLFile(baseDir, path)
		diags = diags.Extend(loadDiags)
		if loadDiags.HasErrors() {
			continue
		}
		realPath := filepath.Join(baseDir, path)

		mod.primaries[i] = f
		mod.Sources[realPath] = f.Bytes
		mod.Files[realPath] = f
	}
	for i, path := range overrides {
		f, loadDiags := p.loadHCLFile(baseDir, path)
		diags = diags.Extend(loadDiags)
		if loadDiags.HasErrors() {
			continue
		}
		realPath := filepath.Join(baseDir, path)

		mod.overrides[i] = f
		mod.Sources[realPath] = f.Bytes
		mod.Files[realPath] = f
	}
	if diags.HasErrors() {
		return mod, diags
	}

	mod.SourceDir = filepath.Join(baseDir, dir)

	buildDiags := mod.build()
	diags = diags.Extend(buildDiags)

	return mod, diags
}

// LoadConfigDirFiles reads the .tf and .tf.json files in the given directory and
// then returns these files as a map of file path.
//
// The difference with LoadConfigDir is that it returns hcl.File instead of
// a single module. This is useful when parsing HCL files in a context outside of
// Terraform.
//
// If a baseDir is passed, the loaded files are assumed to be loaded from that
// directory.
func (p *Parser) LoadConfigDirFiles(baseDir, dir string) (map[string]*hcl.File, hcl.Diagnostics) {
	primaries, overrides, diags := p.configDirFiles(baseDir, dir)
	if diags.HasErrors() {
		return map[string]*hcl.File{}, diags
	}

	files := map[string]*hcl.File{}

	for _, path := range primaries {
		f, loadDiags := p.loadHCLFile(baseDir, path)
		diags = diags.Extend(loadDiags)
		if loadDiags.HasErrors() {
			continue
		}
		files[filepath.Join(baseDir, path)] = f
	}
	for _, path := range overrides {
		f, loadDiags := p.loadHCLFile(baseDir, path)
		diags = diags.Extend(loadDiags)
		if loadDiags.HasErrors() {
			continue
		}
		files[filepath.Join(baseDir, path)] = f
	}

	return files, diags
}

// LoadValuesFile reads the file at the given path and parses it as a "values
// file", which is an HCL config file whose top-level attributes are treated
// as arbitrary key.value pairs.
//
// If the file cannot be read -- for example, if it does not exist -- then
// a nil map will be returned along with error diagnostics. Callers may wish
// to disregard the returned diagnostics in this case and instead generate
// their own error message(s) with additional context.
//
// If the returned diagnostics has errors when a non-nil map is returned
// then the map may be incomplete but should be valid enough for careful
// static analysis.
//
// If a baseDir is passed, the loaded file is assumed to be loaded from that
// directory.
func (p *Parser) LoadValuesFile(baseDir, path string) (map[string]cty.Value, hcl.Diagnostics) {
	f, diags := p.loadHCLFile(baseDir, path)
	if diags.HasErrors() {
		return nil, diags
	}

	vals := make(map[string]cty.Value)
	if f == nil || f.Body == nil {
		return vals, diags
	}

	attrs, attrDiags := f.Body.JustAttributes()
	diags = diags.Extend(attrDiags)
	if attrs == nil {
		return vals, diags
	}

	for name, attr := range attrs {
		val, valDiags := attr.Expr.Value(nil)
		diags = diags.Extend(valDiags)
		vals[name] = val
	}

	return vals, diags
}

func (p *Parser) loadHCLFile(baseDir, path string) (*hcl.File, hcl.Diagnostics) {
	src, err := p.fs.ReadFile(path)

	realPath := filepath.Join(baseDir, path)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Failed to read file",
					Subject:  &hcl.Range{},
					Detail:   fmt.Sprintf("The file %q does not exist.", realPath),
				},
			}
		}
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read file",
				Subject:  &hcl.Range{},
				Detail:   fmt.Sprintf("The file %q could not be read.", realPath),
			},
		}
	}

	switch {
	case strings.HasSuffix(path, ".json"):
		return p.p.ParseJSON(src, realPath)
	default:
		return p.p.ParseHCL(src, realPath)
	}
}

// Sources returns a map of the cached source buffers for all files that
// have been loaded through this parser, with source filenames (as requested
// when each file was opened) as the keys.
func (p *Parser) Sources() map[string][]byte {
	return p.p.Sources()
}

// Files returns a map of the cached HCL file objects for all files that
// have been loaded through this parser, with source filenames (as requested
// when each file was opened) as the keys.
func (p *Parser) Files() map[string]*hcl.File {
	return p.p.Files()
}

func (p *Parser) configDirFiles(baseDir, dir string) (primary, override []string, diags hcl.Diagnostics) {
	infos, err := p.fs.ReadDir(dir)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to read module directory",
			Detail:   fmt.Sprintf("Module directory %s does not exist or cannot be read.", filepath.Join(baseDir, dir)),
		})
		return
	}

	for _, info := range infos {
		if info.IsDir() {
			// We only care about files
			continue
		}

		name := info.Name()
		ext := configFileExt(name)
		if ext == "" || isIgnoredFile(name) {
			continue
		}

		baseName := name[:len(name)-len(ext)] // strip extension
		isOverride := baseName == "override" || strings.HasSuffix(baseName, "_override")

		fullPath := filepath.Join(dir, name)
		if isOverride {
			override = append(override, fullPath)
		} else {
			primary = append(primary, fullPath)
		}
	}

	return
}

func (p *Parser) autoLoadValuesDirFiles(baseDir, dir string) (files []string, diags hcl.Diagnostics) {
	infos, err := p.fs.ReadDir(dir)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to read module directory",
			Detail:   fmt.Sprintf("Module directory %s does not exist or cannot be read.", filepath.Join(baseDir, dir)),
		})
		return nil, diags
	}

	for _, info := range infos {
		if info.IsDir() {
			// We only care about files
			continue
		}

		name := info.Name()
		if !isAutoVarFile(name) {
			continue
		}

		fullPath := filepath.Join(dir, name)
		files = append(files, fullPath)
	}
	// The files should be sorted alphabetically. This is equivalent to priority.
	sort.Strings(files)

	return
}

// configFileExt returns the Terraform configuration extension of the given
// path, or a blank string if it is not a recognized extension.
func configFileExt(path string) string {
	if strings.HasSuffix(path, ".tf") {
		return ".tf"
	} else if strings.HasSuffix(path, ".tf.json") {
		return ".tf.json"
	} else {
		return ""
	}
}

// isAutoVarFile determines if the file ends with .auto.tfvars or .auto.tfvars.json
func isAutoVarFile(path string) bool {
	return strings.HasSuffix(path, ".auto.tfvars") ||
		strings.HasSuffix(path, ".auto.tfvars.json")
}

// isIgnoredFile returns true if the given filename (which must not have a
// directory path ahead of it) should be ignored as e.g. an editor swap file.
func isIgnoredFile(name string) bool {
	return strings.HasPrefix(name, ".") || // Unix-like hidden files
		strings.HasSuffix(name, "~") || // vim
		strings.HasPrefix(name, "#") && strings.HasSuffix(name, "#") // emacs
}
