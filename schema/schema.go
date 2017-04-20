package schema

import (
	"sort"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hcl/hcl/token"
)

type Template struct {
	File      string
	Resources []*Resource
}

func Make(files map[string][]byte) ([]*Template, error) {
	var templates []*Template
	var err error
	plains := map[string][]byte{}
	overrideFiles := []string{}
	overrides := map[string][]byte{}

	for file := range files {
		if file == "override.tf" || strings.HasSuffix(file, "_override.tf") {
			overrideFiles = append(overrideFiles, file)
		} else {
			plains[file] = files[file]
		}
	}
	// Override files are loaded last in alphabetical order.
	sort.Strings(overrideFiles)
	for _, file := range overrideFiles {
		overrides[file] = files[file]
	}

	if templates, err = appendTemplates(templates, plains, false); err != nil {
		return nil, err
	}
	if templates, err = appendTemplates(templates, overrides, true); err != nil {
		return nil, err
	}

	return templates, nil
}

func (t *Template) Find(query ...string) []*Resource {
	resources := []*Resource{}
	var provider, providerType, id string
	for i, attr := range query {
		switch i {
		case 0:
			provider = attr
		case 1:
			providerType = attr
		case 2:
			id = attr
		default:
			// noop
		}
	}

	switch provider {
	case "resource":
		if providerType != "" {
			for _, resource := range t.Resources {
				if id != "" {
					if resource.Type == providerType && resource.Id == id {
						resources = append(resources, resource)
					}
				} else {
					if resource.Type == providerType {
						resources = append(resources, resource)
					}
				}
			}
		} else {
			resources = t.Resources
		}
		return resources
	default:
		return resources
	}
}

func appendTemplates(templates []*Template, files map[string][]byte, override bool) ([]*Template, error) {
	for file, body := range files {
		template := &Template{
			File:      file,
			Resources: []*Resource{},
		}
		var ret map[string]map[string]interface{}
		root, err := parser.Parse(body)
		if err != nil {
			return nil, err
		}
		if err := hcl.DecodeObject(&ret, root); err != nil {
			return nil, err
		}

		for resourceType, typeResources := range ret["resource"] {
			for _, typeResource := range typeResources.([]map[string]interface{}) {
				for key, attrs := range typeResource {
					var newResource bool = true
					var resourceItem *ast.ObjectItem = root.Node.(*ast.ObjectList).Filter("resource", resourceType, key).Items[0]
					var resourcePos token.Pos = resourceItem.Val.Pos()
					resourcePos.Filename = file
					resource := &Resource{
						File:  file,
						Type:  resourceType,
						Id:    key,
						Pos:   resourcePos,
						Attrs: map[string]*Attribute{},
					}

					if override {
						for _, temp := range templates {
							if res := temp.Find("resource", resourceType, key); len(res) != 0 {
								resource = res[0]
								newResource = false
								break
							}
						}
					}

					for _, attr := range attrs.([]map[string]interface{}) {
						for k := range attr {
							for _, attrToken := range resourceItem.Val.(*ast.ObjectType).List.Filter(k).Items {
								if resource.Attrs[k] == nil || override {
									resource.Attrs[k] = &Attribute{}
								}
								// The case of multiple specifiable keys such as `ebs_block_device`.
								resource.Attrs[k].Vals = append(resource.Attrs[k].Vals, getToken(file, attrToken.Val))
								pos := attrToken.Val.Pos()
								pos.Filename = file
								resource.Attrs[k].Poses = append(resource.Attrs[k].Poses, pos)
							}
						}
					}

					if newResource {
						template.Resources = append(template.Resources, resource)
					}
				}
			}
		}

		if !override {
			templates = append(templates, template)
		}
	}

	return templates, nil
}

func getToken(file string, node interface{}) interface{} {
	// attr = "literal"
	if v, ok := node.(*ast.LiteralType); ok {
		v.Token.Pos.Filename = file
		// token.Token{}
		return v.Token
	}
	// attr = ["elem1", "elem2"]
	if v, ok := node.(*ast.ListType); ok {
		tokens := []interface{}{}
		for _, childNode := range v.List {
			tokens = append(tokens, getToken(file, childNode))
		}
		// []token.Token{}
		return tokens
	}
	// attr = {
	//   "key" = "value"
	// }
	if v, ok := node.(*ast.ObjectType); ok {
		tokenMap := map[string]interface{}{}
		for _, item := range v.List.Items {
			tokenMap[item.Keys[0].Token.Text] = getToken(file, item.Val)
		}
		// map[string]token.Token{}
		return tokenMap
	}
	// Unexpected node pattern
	return token.Token{}
}
