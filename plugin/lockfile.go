package plugin

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

// LockfileVersion is the lockfile format version that this TFLint writes.
//
// The version must be bumped when making changes that older TFLint cannot
// safely ignore, such as structural changes or additions related to verification.
// Attributes that older TFLint can safely ignore, and that newer TFLint can
// live without, may be added without bumping the version.
//
// When bumping the version, it is recommended to release the ability to read
// the new version first, and start writing it in a later release.
const LockfileVersion = 1

// hashSchemeH1 is the prefix for the sha256 hash of a plugin binary.
// Note that it is not the same as Terraform's "h1:" (dirhash).
const hashSchemeH1 = "h1:"

const lockfileHeader = `# This file is maintained automatically by "tflint --init".
# Manual edits may be lost in future updates.
`

var (
	platformPattern = regexp.MustCompile(`^[a-z0-9]+_[a-z0-9]+$`)
	h1HashPattern   = regexp.MustCompile(`^h1:[0-9a-f]{64}$`)
)

// Lockfile represents a dependency lockfile (.tflint.lock.hcl).
// It records the trusted plugins to prevent installing anything other
// than the plugins that were trusted when they were first installed.
type Lockfile struct {
	// Version is the lockfile format version.
	Version int
	// Plugins are locked plugins keyed by source.
	Plugins map[string]*LockedPlugin

	path string
	// raw is the original file content. Used to avoid unnecessary writes.
	raw []byte
}

// LockedPlugin is a lock entry for a plugin.
type LockedPlugin struct {
	// Source is the plugin source, e.g. github.com/terraform-linters/tflint-ruleset-aws
	Source string
	// Version is the plugin version, e.g. 0.35.0
	Version string
	// RepositoryID is the ID of the source repository on GitHub.
	// It changes when the repository is deleted and recreated with the same name.
	RepositoryID int64
	// Hashes are hashes of the plugin binary keyed by platform (e.g. linux_amd64).
	// The value is like "h1:<sha256 hex>".
	Hashes map[string]string
}

// LockfilePath returns the lockfile path derived from the config file path.
// The lockfile is placed next to the config file, and its name is derived
// by replacing the ".hcl" or ".json" extension with ".lock.hcl".
//
//   - .tflint.hcl -> .tflint.lock.hcl
//   - .tflint.json -> .tflint.lock.hcl
//   - configs/aws.hcl -> configs/aws.lock.hcl
//   - tflintrc -> tflintrc.lock.hcl
//
// Returns an empty string if the config path is empty, which means
// the default config is used and no lockfile is available.
func LockfilePath(configPath string) string {
	if configPath == "" {
		return ""
	}
	switch filepath.Ext(configPath) {
	case ".hcl", ".json":
		configPath = strings.TrimSuffix(configPath, filepath.Ext(configPath))
	}
	return configPath + ".lock.hcl"
}

// NewLockfile returns an empty lockfile for the given path.
func NewLockfile(path string) *Lockfile {
	return &Lockfile{
		Version: LockfileVersion,
		Plugins: map[string]*LockedPlugin{},
		path:    path,
	}
}

// LoadLockfile loads the lockfile from the given path.
// If the file does not exist, nil is returned without error.
func LoadLockfile(path string) (*Lockfile, error) {
	if path == "" {
		return nil, nil
	}

	src, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to read lockfile %s; %w", path, err)
	}

	lockfile, diags := parseLockfile(src, path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("Failed to parse lockfile %s; %w", path, diags)
	}
	return lockfile, nil
}

var lockfileVersionSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "lockfile_version", Required: true},
	},
}

var lockfileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "plugin", LabelNames: []string{"source"}},
	},
}

var lockedPluginSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "version", Required: true},
		{Name: "repository_id", Required: true},
		{Name: "hashes", Required: true},
	},
}

func parseLockfile(src []byte, filename string) (*Lockfile, hcl.Diagnostics) {
	file, diags := hclsyntax.ParseConfig(src, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, diags
	}

	// Check the lockfile version first because the structure may differ in other versions.
	// Unknown attributes and blocks are ignored to accept additions in the same version.
	content, _, diags := file.Body.PartialContent(lockfileVersionSchema)
	if diags.HasErrors() {
		return nil, diags
	}
	attr := content.Attributes["lockfile_version"]
	var version int
	if diags := gohcl.DecodeExpression(attr.Expr, nil, &version); diags.HasErrors() {
		return nil, diags
	}
	if version > LockfileVersion {
		return nil, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Unsupported lockfile version",
			Detail:   fmt.Sprintf("The lockfile version %d is not supported by this version of TFLint (supported up to %d). Please upgrade TFLint.", version, LockfileVersion),
			Subject:  attr.Expr.Range().Ptr(),
		}}
	}
	if version < 1 {
		return nil, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Invalid lockfile version",
			Detail:   fmt.Sprintf("The lockfile version must be a positive integer, but got %d.", version),
			Subject:  attr.Expr.Range().Ptr(),
		}}
	}

	lockfile := &Lockfile{
		Version: version,
		Plugins: map[string]*LockedPlugin{},
		path:    filename,
		raw:     src,
	}

	content, _, diags = file.Body.PartialContent(lockfileSchema)
	if diags.HasErrors() {
		return nil, diags
	}
	for _, block := range content.Blocks {
		source := block.Labels[0]
		if _, exists := lockfile.Plugins[source]; exists {
			return nil, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Duplicate plugin entry",
				Detail:   fmt.Sprintf(`The plugin "%s" is locked more than once.`, source),
				Subject:  block.DefRange.Ptr(),
			}}
		}

		plugin, diags := decodeLockedPlugin(source, block)
		if diags.HasErrors() {
			return nil, diags
		}
		lockfile.Plugins[source] = plugin
	}

	return lockfile, nil
}

func decodeLockedPlugin(source string, block *hcl.Block) (*LockedPlugin, hcl.Diagnostics) {
	content, _, diags := block.Body.PartialContent(lockedPluginSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	plugin := &LockedPlugin{Source: source}
	if diags := gohcl.DecodeExpression(content.Attributes["version"].Expr, nil, &plugin.Version); diags.HasErrors() {
		return nil, diags
	}
	if diags := gohcl.DecodeExpression(content.Attributes["repository_id"].Expr, nil, &plugin.RepositoryID); diags.HasErrors() {
		return nil, diags
	}

	hashesAttr := content.Attributes["hashes"]
	if diags := gohcl.DecodeExpression(hashesAttr.Expr, nil, &plugin.Hashes); diags.HasErrors() {
		return nil, diags
	}
	for platform, hash := range plugin.Hashes {
		if !platformPattern.MatchString(platform) {
			return nil, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Invalid platform",
				Detail:   fmt.Sprintf(`The platform "%s" must be in the form of "<os>_<arch>".`, platform),
				Subject:  hashesAttr.Expr.Range().Ptr(),
			}}
		}
		if !h1HashPattern.MatchString(hash) {
			return nil, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Invalid hash",
				Detail:   fmt.Sprintf(`The hash "%s" for %s must be in the form of "h1:<sha256 hex>".`, hash, platform),
				Subject:  hashesAttr.Expr.Range().Ptr(),
			}}
		}
	}

	return plugin, nil
}

// Path returns the lockfile path.
func (l *Lockfile) Path() string {
	return l.path
}

// Bytes returns the lockfile content in a deterministic format.
func (l *Lockfile) Bytes() []byte {
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.SetAttributeValue("lockfile_version", cty.NumberIntVal(int64(l.Version)))

	for _, source := range slices.Sorted(maps.Keys(l.Plugins)) {
		plugin := l.Plugins[source]

		body.AppendNewline()
		block := body.AppendNewBlock("plugin", []string{source}).Body()
		block.SetAttributeValue("version", cty.StringVal(plugin.Version))
		block.SetAttributeValue("repository_id", cty.NumberIntVal(plugin.RepositoryID))

		hashes := map[string]cty.Value{}
		for platform, hash := range plugin.Hashes {
			hashes[platform] = cty.StringVal(hash)
		}
		if len(hashes) == 0 {
			block.SetAttributeValue("hashes", cty.EmptyObjectVal)
		} else {
			block.SetAttributeValue("hashes", cty.ObjectVal(hashes))
		}
	}

	return append([]byte(lockfileHeader+"\n"), hclwrite.Format(f.Bytes())...)
}

// Save writes the lockfile only if the content has changed.
// Returns true if the file was written.
func (l *Lockfile) Save() (bool, error) {
	out := l.Bytes()
	if l.raw != nil && bytes.Equal(out, l.raw) {
		return false, nil
	}

	if err := os.WriteFile(l.path, out, 0o644); err != nil {
		return false, fmt.Errorf("Failed to write lockfile %s; %w", l.path, err)
	}
	l.raw = out
	return true, nil
}

// checkInstalledPlugin checks the consistency between the installed plugin and the lockfile
// before launching it. If the version does not match the lock entry, it returns an error.
// If the plugin is not locked, it returns a warning message because plugins installed
// before introducing the lockfile are not recorded until they are reinstalled.
func (l *Lockfile) checkInstalledPlugin(cfg *tflint.PluginConfig, pluginPath string) (string, error) {
	entry, exists := l.Plugins[cfg.Source]
	if !exists {
		return fmt.Sprintf(`Plugin "%s" (source: %s, version: %s) is not recorded in the lockfile %s, so it is not verified. To record it, remove %s and run "tflint --init"`, cfg.Name, cfg.Source, cfg.Version, l.path, pluginPath), nil
	}
	if entry.Version != cfg.Version {
		return "", fmt.Errorf(`Plugin "%s" version %s does not match the lockfile %s (locked: %s). Did you run "tflint --init"?`, cfg.Name, cfg.Version, l.path, entry.Version)
	}
	return "", nil
}

// currentPlatform returns the platform name of the running TFLint, e.g. linux_amd64.
func currentPlatform() string {
	return fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
}

// hashFile returns the "h1:" hash of the given file.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hashSchemeH1 + hex.EncodeToString(h.Sum(nil)), nil
}
