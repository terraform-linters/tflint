package schema

import "github.com/hashicorp/hcl/hcl/token"

type Source struct {
	File  string
	Pos   token.Pos
	Attrs map[string]*Attribute
}

type Attribute struct {
	Poses []token.Pos
	Vals  []interface{}
}

func (s *Source) GetToken(name string) (token.Token, bool) {
	if s.Attrs[name] != nil {
		token, ok := s.Attrs[name].Vals[0].(token.Token)
		return token, ok
	}
	return token.Token{}, false
}

func (s *Source) GetListToken(name string) ([]token.Token, bool) {
	if s.Attrs[name] != nil {
		val, ok := s.Attrs[name].Vals[0].([]interface{})
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

func (s *Source) GetMapToken(name string) (map[string]token.Token, bool) {
	var tokens map[string]token.Token = map[string]token.Token{}
	if s.Attrs[name] != nil {
		cval, ok := s.Attrs[name].Vals[0].(map[string]interface{})
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

func (s *Source) GetAllMapTokens(name string) ([]map[string]token.Token, bool) {
	var tokens []map[string]token.Token = []map[string]token.Token{}
	if s.Attrs[name] != nil {
		for _, val := range s.Attrs[name].Vals {
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
