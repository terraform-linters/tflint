package schema

type Provider struct {
	*Source
	Id           string
	Type 				 string
	Templates    []*Template
	EvalConfig   hil.EvalConfig
}

func newProvider(fileName string, pos token.Pos, providerType string, providerId string) *Provider {
	return &Provider{
		Id: providerId,
		Type: providerType,
		Source: &Source{
			File:  fileName,
			Pos:   pos,
			Attrs: map[string]*Attribute{},
		},
	}
}