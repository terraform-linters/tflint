package schema

import "github.com/hashicorp/hcl/hcl/token"

type Resource struct {
	*Source
	Type string
	Id   string
}

func NewResource(fileName string, pos token.Pos, resourceType string, resourceId string) *Resource {
	return &Resource{
		Type: resourceType,
		Id:   resourceId,
		Source: &Source{
			File:  fileName,
			Pos:   pos,
			Attrs: map[string]*Attribute{},
		},
	}
}
