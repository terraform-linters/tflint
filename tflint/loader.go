package tflint

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
)

// Loader is a wrapper of Terraform's configload.Loader
type Loader struct {
	loader               *configload.Loader
	moduleSourceVersions map[string][]*version.Version
	moduleManifest       map[string]*moduleManifest
}

type moduleManifest struct {
	Key        string           `json:"Key"`
	Source     string           `json:"Source"`
	Version    *version.Version `json:"-"`
	VersionStr string           `json:"Version,omitempty"`
	Dir        string           `json:"Dir"`
	Root       string           `json:"Root"`
}

type moduleManifestFile struct {
	Modules []*moduleManifest `json:"Modules"`
}

// NewLoader returns a loader with module manifests
func NewLoader() (*Loader, error) {
	loader, err := configload.NewLoader(&configload.Config{
		ModulesDir: getTFModuleDir(),
	})
	if err != nil {
		return nil, err
	}

	l := &Loader{
		loader:               loader,
		moduleSourceVersions: map[string][]*version.Version{},
		moduleManifest:       map[string]*moduleManifest{},
	}

	if _, err := os.Stat(getTFModuleManifestPath()); !os.IsNotExist(err) {
		if err := l.initializeModuleManifest(); err != nil {
			return nil, err
		}
	}

	return l, nil
}

// LoadConfig loads Terraform's configurations
func (l *Loader) LoadConfig() (*configs.Config, error) {
	rootMod, diags := l.loader.Parser().LoadConfigDir(".")
	if diags.HasErrors() {
		return nil, diags
	}

	cfg, diags := configs.BuildConfig(rootMod, l.moduleWalkerLegacy())
	if !diags.HasErrors() {
		return cfg, nil
	}

	cfg, diags = configs.BuildConfig(rootMod, l.moduleWalkerV0_10_6())
	if !diags.HasErrors() {
		return cfg, nil
	}

	cfg, diags = configs.BuildConfig(rootMod, l.moduleWalkerV0_10_7V0_10_8())
	if !diags.HasErrors() {
		return cfg, nil
	}

	cfg, diags = configs.BuildConfig(rootMod, l.moduleWalkerV0_11_0V0_11_7())
	if !diags.HasErrors() {
		return cfg, nil
	}

	return nil, diags
}

func (l *Loader) moduleWalkerLegacy() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		mod, diags := l.loader.Parser().LoadConfigDir(makeModuleDirFromKey("root." + req.Name + "-" + req.SourceAddr))
		return mod, nil, diags
	})
}

func (l *Loader) moduleWalkerV0_10_6() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		mod, diags := l.loader.Parser().LoadConfigDir(makeModuleDirFromKey("module." + req.Name + "-" + req.SourceAddr))
		return mod, nil, diags
	})
}

func (l *Loader) moduleWalkerV0_10_7V0_10_8() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		mod, diags := l.loader.Parser().LoadConfigDir(makeModuleDirFromKey("0.root." + req.Name + "-" + req.SourceAddr))
		return mod, nil, diags
	})
}

func (l *Loader) moduleWalkerV0_11_0V0_11_7() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		path := append(buildParentModulePathTree([]string{}, req.Parent), l.getModulePath(req))
		key := "1." + strings.Join(path, "|")

		record, ok := l.moduleManifest[key]
		if !ok {
			return nil, nil, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("`%s` module is not found. Did you run `terraform init`?", req.Name),
					Detail:   fmt.Sprintf("Failed to search by `%s` key.", key),
					Subject:  &req.CallRange,
				},
			}
		}

		dir := record.Dir
		if record.Root != "" {
			dir = filepath.Join(dir, record.Root)
		}

		mod, diags := l.loader.Parser().LoadConfigDir(dir)
		return mod, record.Version, diags
	})
}

func (l *Loader) initializeModuleManifest() error {
	file, err := ioutil.ReadFile(getTFModuleManifestPath())
	if err != nil {
		return err
	}

	var manifestFile moduleManifestFile
	err = json.Unmarshal(file, &manifestFile)
	if err != nil {
		return err
	}

	for _, m := range manifestFile.Modules {
		if m.VersionStr != "" {
			m.Version, err = version.NewVersion(m.VersionStr)
			if err != nil {
				return err
			}
			l.moduleSourceVersions[m.Source] = append(l.moduleSourceVersions[m.Source], m.Version)
		}
		l.moduleManifest[m.Key] = m
	}

	return nil
}

func makeModuleDirFromKey(key string) string {
	sum := md5.Sum([]byte(key))
	return filepath.Join(getTFModuleDir(), hex.EncodeToString(sum[:]))
}

func buildParentModulePathTree(path []string, cfg *configs.Config) []string {
	if cfg.Path.IsRoot() {
		// @see https://github.com/golang/go/wiki/SliceTricks#reversing
		for i := len(path)/2 - 1; i >= 0; i-- {
			opp := len(path) - 1 - i
			path[i], path[opp] = path[opp], path[i]
		}
		return path
	}

	_, call := cfg.Path.Call()
	key := call.Name
	if cfg.Version != nil {
		key += "#" + cfg.Version.String()
	}
	key += ";" + cfg.SourceAddr
	path = append(path, key)

	return buildParentModulePathTree(path, cfg.Parent)
}

func (l *Loader) getModulePath(req *configs.ModuleRequest) string {
	key := req.Name + ";" + req.SourceAddr
	if len(req.VersionConstraint.Required) > 0 {
		sourceVersions := l.moduleSourceVersions[req.SourceAddr]

		var latest *version.Version
		for _, v := range sourceVersions {
			if req.VersionConstraint.Required.Check(v) && (latest == nil || v.GreaterThan(latest)) {
				latest = v
			}
		}
		key += "." + latest.String()
	}

	return key
}
