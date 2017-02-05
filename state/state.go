package state

const LocalStatePath string = "terraform.tfstate"
const RemoteStatePath string = ".terraform/terraform.tfstate"

type TFState struct {
	Modules []*Module `json:"modules"`
}

type Module struct {
	Resources map[string]*Resource `json:"resources"`
}

type Resource struct {
	Type         string    `json:"type"`
	Dependencies []string  `json:"depends_on"`
	Primary      *Instance `json:"primary"`
	Provider     string    `json:"provider"`
}

type Instance struct {
	ID         string            `json:"id"`
	Attributes map[string]string `json:"attributes"`
}

func (s *TFState) Exists(resourceType string, id string) bool {
	for _, module := range s.Modules {
		if _, ok := module.Resources[resourceType+"."+id]; ok {
			return true
		}
	}
	return false
}
