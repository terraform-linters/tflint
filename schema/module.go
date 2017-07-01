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

	if _, err := os.Stat(m.path()); err != nil {
		return fmt.Errorf("ERROR: module `%s` not found. Did you run `terraform get`?", m.ModuleSource)
	}

	filePaths, err := filepath.Glob(m.path() + "/*.tf")
	if err != nil {
		return err
	}

	for _, filePath := range filePaths {
		b, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("ERROR: Cannot open file %s", filePath)
		}

		fileName := strings.Replace(filePath, m.path(), "", 1)
		fileKey := m.ModuleSource + fileName
		files[fileKey] = b
	}

	if m.Templates, err = Make(files); err != nil {
		return err
	}

	return nil
}

func (m *Module) path() string {
	base := "root." + m.Id + "-" + m.ModuleSource
	sum := md5.Sum([]byte(base)) // #nosec
	return ".terraform/modules/" + hex.EncodeToString(sum[:])
}
