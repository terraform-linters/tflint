package terraform

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform/addrs"
)

func moduleManifestPath() string {
	return filepath.Join(dataDir(), "modules", "modules.json")
}

// moduleMgr is a fork of configload.moduleMgr. It manages the installation
// state for modules. Unlike Terraform, it does not install from the registry
// and is read-only.
type moduleMgr struct {
	fs       afero.Afero
	manifest moduleManifest
}

// moduleRecord is a fork of modsdir.Record. This describes the structure of
// the manifest file which is usually placed in .terraform/modules/modules.json.
type moduleRecord struct {
	Key        string           `json:"Key"`
	Source     string           `json:"Source"`
	Version    *version.Version `json:"-"`
	VersionStr string           `json:"Version,omitempty"`
	Dir        string           `json:"Dir"`
}

type moduleManifestFile struct {
	Records []*moduleRecord `json:"Modules"`
}

type moduleManifest map[string]*moduleRecord

func (m moduleManifest) moduleKey(path addrs.Module) string {
	if len(path) == 0 {
		return ""
	}
	return strings.Join([]string(path), ".")
}

func (l *moduleMgr) readModuleManifest() error {
	r, err := l.fs.Open(moduleManifestPath())
	if err != nil {
		if os.IsNotExist(err) {
			// We'll treat a missing file as an empty manifest
			l.manifest = make(moduleManifest)
			return nil
		}
		return err
	}

	log.Print("[INFO] Module manifest file found. Initializing...")

	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var read moduleManifestFile
	err = json.Unmarshal(src, &read)
	if err != nil {
		return fmt.Errorf("error unmarshalling module manifest file: %v", err)
	}

	for _, record := range read.Records {
		l.manifest[record.Key] = record
	}

	return nil
}
