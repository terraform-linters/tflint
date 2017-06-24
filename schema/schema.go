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
	Modules   []*Module
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
			if templates, err = appendTemplates(templates, file, files[file], false); err != nil {
				return nil, err
			}
		}
	}
	// Override files are loaded last in alphabetical order.
	sort.Strings(overrideFiles)
	for _, file := range overrideFiles {
		overrides[file] = files[file]
		if templates, err = appendTemplates(templates, file, files[file], true); err != nil {
			return nil, err
		}
	}

	return templates, nil
}

func (t *Template) FindResources(query ...string) []*Resource {
	resources := []*Resource{}
	var resourceType, id string
	for i, attr := range query {
		switch i {
		case 0:
			resourceType = attr
		case 1:
			id = attr
		default:
			// noop
		}
	}

	if resourceType != "" {
		for _, resource := range t.Resources {
			if id != "" {
				if resource.Type == resourceType && resource.Id == id {
					resources = append(resources, resource)
				}
			} else {
				if resource.Type == resourceType {
					resources = append(resources, resource)
				}
			}
		}
	} else {
		resources = t.Resources
	}
	return resources
}

func (t *Template) FindModules(id string) []*Module {
	modules := []*Module{}

	for _, module := range t.Modules {
		if module.Id == id {
			modules = append(modules, module)
		}
	}

	return modules
}

func appendTemplates(templates []*Template, file string, body []byte, override bool) ([]*Template, error) {
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
				var new bool = true
				var resourceItem *ast.ObjectItem = root.Node.(*ast.ObjectList).Filter("resource", resourceType, key).Items[0]
				var resourcePos token.Pos = resourceItem.Val.Pos()
				resourcePos.Filename = file
				resource := newResource(file, resourcePos, resourceType, key)

				if override {
					for _, temp := range templates {
						if res := temp.FindResources(resourceType, key); len(res) != 0 {
							resource = res[0]
							new = false
							break
						}
					}
				}
				resource.setAttrs(attrs, resourceItem, file, override)

				if new {
					template.Resources = append(template.Resources, resource)
				}
			}
		}
	}

	for moduleId, attrs := range ret["module"] {
		var new bool = true
		var moduleItem *ast.ObjectItem = root.Node.(*ast.ObjectList).Filter("module", moduleId).Items[0]
		var modulePos token.Pos = moduleItem.Val.Pos()
		modulePos.Filename = file
		module := newModule(file, modulePos, moduleId)

		if override {
			for _, temp := range templates {
				if mod := temp.FindModules(moduleId); len(mod) != 0 {
					module = mod[0]
					new = false
					break
				}
			}
		}
		module.setAttrs(attrs, moduleItem, file, override)

		if new {
			template.Modules = append(template.Modules, module)
		}
	}

	if !override {
		templates = append(templates, template)
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
