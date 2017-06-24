package schema

import "github.com/hashicorp/hcl/hcl/token"

type Module struct {
	*Source
	Id string
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
