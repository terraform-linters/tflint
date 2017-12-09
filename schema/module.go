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
	modulePath, err := m.path()
	if err != nil {
		return err
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

func (m *Module) path() (string, error) {
	// before v0.10.5
	base := m.pathname("root." + m.Id + "-" + m.ModuleSource)
	if _, err := os.Stat(base); err == nil {
		return base, nil
	}
	// for v0.10.6
	base = m.pathname("module." + m.Id + "-" + m.ModuleSource)
	if _, err := os.Stat(base); err == nil {
		return base, nil
	}
	// for v0.10.7 and 0.10.8
	base = m.pathname("0.root." + m.Id + "-" + m.ModuleSource)
	if _, err := os.Stat(base); err == nil {
		return base, nil
	}
	// after v0.11.0
	// XXX: Unfortunately, this format can not resolve the path of the module that specifies the version.
	base = m.pathname("1." + m.Id + ";" + m.ModuleSource)
	if _, err := os.Stat(base); err == nil {
		return base, nil
	}

	return "", fmt.Errorf("ERROR: module `%s` not found. Did you run `terraform get`?", m.ModuleSource)
}

func (m *Module) pathname(base string) string {
	sum := md5.Sum([]byte(base)) // #nosec
	return ".terraform/modules/" + hex.EncodeToString(sum[:])
}
