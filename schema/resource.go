package schema

import "github.com/hashicorp/hcl/hcl/token"

type Resource struct {
	File  string
	Type  string
	Id    string
	Pos   token.Pos
	Attrs map[string]*Attribute
}

type Attribute struct {
	Poses []token.Pos
	Vals  []interface{}
}

func (r *Resource) GetToken(name string) (token.Token, bool) {
	if r.Attrs[name] != nil {
		token, ok := r.Attrs[name].Vals[0].(token.Token)
		return token, ok
	}
	return token.Token{}, false
}

func (r *Resource) GetListToken(name string) ([]token.Token, bool) {
	if r.Attrs[name] != nil {
		val, ok := r.Attrs[name].Vals[0].([]interface{})
		if !ok {
			return []token.Token{}, false
		}

		tokens := []token.Token{}
		for _, v := range val {
			t, ok := v.(token.Token)
			if !ok {
				return []token.Token{}, false
			}
			tokens = append(tokens, t)
		}
		return tokens, true
	}
	return []token.Token{}, false
}

func (r *Resource) GetMapToken(name string) (map[string]token.Token, bool) {
	var tokens map[string]token.Token = map[string]token.Token{}
	if r.Attrs[name] != nil {
		cval, ok := r.Attrs[name].Vals[0].(map[string]interface{})
		if !ok {
			return map[string]token.Token{}, false
		}

		for k, v := range cval {
			cv, ok := v.(token.Token)
			if !ok {
				return map[string]token.Token{}, false
			}
			tokens[k] = cv
		}
		return tokens, true
	}
	return map[string]token.Token{}, false
}

func (r *Resource) GetAllMapTokens(name string) ([]map[string]token.Token, bool) {
	var tokens []map[string]token.Token = []map[string]token.Token{}
	if r.Attrs[name] != nil {
		for _, val := range r.Attrs[name].Vals {
			cval, ok := val.(map[string]interface{})
			if !ok {
				return []map[string]token.Token{}, false
			}

			var tokenMap map[string]token.Token = map[string]token.Token{}
			for k, v := range cval {
				cv, ok := v.(token.Token)
				if !ok {
					return []map[string]token.Token{}, false
				}
				tokenMap[k] = cv
			}
			tokens = append(tokens, tokenMap)
		}
		return tokens, true
	}
	return []map[string]token.Token{}, false
}
