package schema

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"strings"

	"fmt"
	"os"

	"path/filepath"

	"github.com/hashicorp/hcl/hcl/token"
	"github.com/hashicorp/hil"
)

type Module struct {
	*Source
	Id           string
	ModuleSource string
	Templates    []*Template
	EvalConfig   hil.EvalConfig
}

func newModule(fileName string, pos token.Pos, moduleId string) *Module {
	return &Module{
		Id: moduleId,
		Source: &Source{
			File:  fileName,
			Pos:   pos,
			Attrs: map[string]*Attribute{},
		},
	}
}

func (m *Module) Load() error {
	files := map[string][]byte{}
	modulePath := m.oldPath()

	if _, err := os.Stat(modulePath); err != nil {
		// Since the digest has changed in Terraform v0.10.6, try to check the new path
		modulePath = m.path()
		if _, err := os.Stat(modulePath); err != nil {
			return fmt.Errorf("ERROR: module `%s` not found. Did you run `terraform get`?", m.ModuleSource)
		}
	}

	filePaths, err := filepath.Glob(modulePath + "/*.tf")
	if err != nil {
		return err
	}

	for _, filePath := range filePaths {
		b, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("ERROR: Cannot open file %s", filePath)
		}

		fileName := strings.Replace(filePath, modulePath, "", 1)
		fileKey := m.ModuleSource + fileName
		files[fileKey] = b
	}

	if m.Templates, err = Make(files); err != nil {
		return err
	}

	return nil
}

// Since the digest has changed in Terraform v0.10.6, this is the path up to v0.10.5
func (m *Module) oldPath() string {
	base := "root." + m.Id + "-" + m.ModuleSource
	sum := md5.Sum([]byte(base)) // #nosec
	return ".terraform/modules/" + hex.EncodeToString(sum[:])
}

func (m *Module) path() string {
	base := "module." + m.Id + "-" + m.ModuleSource
	sum := md5.Sum([]byte(base)) // #nosec
	return ".terraform/modules/" + hex.EncodeToString(sum[:])
}
