package terraform

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

// Parser is the main interface to read configuration files and other related
// files from disk.
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
// will simply return an empty module in that case. Callers should first call
// Parser.IsConfigDir if they wish to recognize that situation.
//
// .tf files are parsed using the HCL native syntax while .tf.json files are
// parsed using the HCL JSON syntax.
func (p *Parser) LoadConfigDir(dir string) (*Module, hcl.Diagnostics) {
	primaries, overrides, diags := p.dirFiles(dir)
	if diags.HasErrors() {
		return nil, diags
	}

	mod := NewEmptyModule()
	mod.primaries = make([]*hcl.File, len(primaries))
	mod.overrides = make([]*hcl.File, len(overrides))

	for i, path := range primaries {
		f, loadDiags := p.loadHCLFile(path)
		diags = diags.Extend(loadDiags)

		mod.primaries[i] = f
		mod.Sources[path] = f.Bytes
		mod.Files[path] = f
	}
	for i, path := range overrides {
		f, loadDiags := p.loadHCLFile(path)
		diags = diags.Extend(loadDiags)

		mod.overrides[i] = f
		mod.Sources[path] = f.Bytes
		mod.Files[path] = f
	}
	if diags.HasErrors() {
		return mod, diags
	}

	mod.SourceDir = dir

	buildDiags := mod.build()
	diags = diags.Extend(buildDiags)

	return mod, diags
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
// This method wraps LoadHCLFile, and so it inherits the syntax selection
// behaviors documented for that method.
func (p *Parser) LoadValuesFile(path string) (map[string]cty.Value, hcl.Diagnostics) {
	f, diags := p.loadHCLFile(path)
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

func (p *Parser) loadHCLFile(path string) (*hcl.File, hcl.Diagnostics) {
	src, err := p.fs.ReadFile(path)

	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The file %q could not be read.", path),
			},
		}
	}

	switch {
	case strings.HasSuffix(path, ".json"):
		return p.p.ParseJSON(src, path)
	default:
		return p.p.ParseHCL(src, path)
	}
}

// Sources returns a map of the cached source buffers for all files that
// have been loaded through this parser, with source filenames (as requested
// when each file was opened) as the keys.
func (p *Parser) Sources() map[string][]byte {
	return p.p.Sources()
}

// ConfigDirFiles returns lists of the primary and override files configuration
// files in the given directory.
//
// If the given directory does not exist or cannot be read, error diagnostics
// are returned. If errors are returned, the resulting lists may be incomplete.
func (p Parser) ConfigDirFiles(dir string) (primary, override []string, diags hcl.Diagnostics) {
	return p.dirFiles(dir)
}

// IsConfigDir determines whether the given path refers to a directory that
// exists and contains at least one Terraform config file (with a .tf or
// .tf.json extension.)
func (p *Parser) IsConfigDir(path string) bool {
	primaryPaths, overridePaths, _ := p.dirFiles(path)
	return (len(primaryPaths) + len(overridePaths)) > 0
}

func (p *Parser) dirFiles(dir string) (primary, override []string, diags hcl.Diagnostics) {
	infos, err := p.fs.ReadDir(dir)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to read module directory",
			Detail:   fmt.Sprintf("Module directory %s does not exist or cannot be read.", dir),
		})
		return
	}

	for _, info := range infos {
		if info.IsDir() {
			// We only care about files
			continue
		}

		name := info.Name()
		ext := fileExt(name)
		if ext == "" || isIgnoredFile(name) {
			continue
		}

		baseName := name[:len(name)-len(ext)] // strip extension
		isOverride := baseName == "override" || strings.HasSuffix(baseName, "_override")

		// TODO: path normalization
		fullPath := filepath.Join(dir, name)
		if isOverride {
			override = append(override, fullPath)
		} else {
			primary = append(primary, fullPath)
		}
	}

	return
}

// fileExt returns the Terraform configuration extension of the given
// path, or a blank string if it is not a recognized extension.
func fileExt(path string) string {
	if strings.HasSuffix(path, ".tf") {
		return ".tf"
	} else if strings.HasSuffix(path, ".tf.json") {
		return ".tf.json"
	} else {
		return ""
	}
}

// isIgnoredFile returns true if the given filename (which must not have a
// directory path ahead of it) should be ignored as e.g. an editor swap file.
func isIgnoredFile(name string) bool {
	return strings.HasPrefix(name, ".") || // Unix-like hidden files
		strings.HasSuffix(name, "~") || // vim
		strings.HasPrefix(name, "#") && strings.HasSuffix(name, "#") // emacs
}
