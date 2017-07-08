package schema

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/hcl/token"
	"github.com/k0kubun/pp"
)

var literalToken = []interface{}{
	token.Token{
		Type: 9,
		Pos: token.Pos{
			Filename: "test.tf",
			Offset:   51,
			Line:     3,
			Column:   19,
		},
		Text: "\"literal value\"",
		JSON: false,
	},
}

var listToken = []interface{}{
	[]interface{}{
		token.Token{
			Type: 9,
			Pos: token.Pos{
				Filename: "test.tf",
				Offset:   208,
				Line:     12,
				Column:   22,
			},
			Text: "\"list1\"",
			JSON: false,
		},
		token.Token{
			Type: 9,
			Pos: token.Pos{
				Filename: "test.tf",
				Offset:   216,
				Line:     12,
				Column:   30,
			},
			Text: "\"list2\"",
			JSON: false,
		},
	},
}

var mapToken = []interface{}{
	map[string]interface{}{
		"Key": token.Token{
			Type: 9,
			Pos: token.Pos{
				Filename: "test.tf",
				Offset:   134,
				Line:     7,
				Column:   12,
			},
			Text: "\"Value\"",
			JSON: false,
		},
	},
}

var multiMapToken = []interface{}{
	map[string]interface{}{
		"Key1": token.Token{
			Type: 9,
			Pos: token.Pos{
				Filename: "test.tf",
				Offset:   134,
				Line:     7,
				Column:   12,
			},
			Text: "\"Value1\"",
			JSON: false,
		},
	},
	map[string]interface{}{
		"Key2": token.Token{
			Type: 9,
			Pos: token.Pos{
				Filename: "test2.tf",
				Offset:   134,
				Line:     7,
				Column:   12,
			},
			Text: "\"Value2\"",
			JSON: false,
		},
	}}

var source = Source{
	Attrs: map[string]*Attribute{
		"literal": {
			Vals: literalToken,
		},
		"list": {
			Vals: listToken,
		},
		"map": {
			Vals: mapToken,
		},
		"multiMap": {
			Vals: multiMapToken,
		},
	},
}

func TestGetToken(t *testing.T) {
	type Result struct {
		Token token.Token
		Ok    bool
	}

	cases := []struct {
		Name   string
		Input  string
		Result Result
	}{
		{
			Name:  "Get literal token",
			Input: "literal",
			Result: Result{
				Token: token.Token{
					Type: 9,
					Pos: token.Pos{
						Filename: "test.tf",
						Offset:   51,
						Line:     3,
						Column:   19,
					},
					Text: "\"literal value\"",
					JSON: false,
				},
				Ok: true,
			},
		},
		{
			Name:  "Get another token",
			Input: "list",
			Result: Result{
				Token: token.Token{},
				Ok:    false,
			},
		},
		{
			Name:  "Get token by not exists key",
			Input: "unknown",
			Result: Result{
				Token: token.Token{},
				Ok:    false,
			},
		},
	}

	for _, tc := range cases {
		token, ok := source.GetToken(tc.Input)
		if !reflect.DeepEqual(token, tc.Result.Token) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(token), pp.Sprint(tc.Result.Token), tc.Name)
		}

		if ok != tc.Result.Ok {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(ok), pp.Sprint(tc.Result.Ok), tc.Name)
		}
	}
}

func TestGetListToken(t *testing.T) {
	type Result struct {
		Token []token.Token
		Ok    bool
	}

	cases := []struct {
		Name   string
		Input  string
		Result Result
	}{
		{
			Name:  "Get list token",
			Input: "list",
			Result: Result{
				Token: []token.Token{
					{
						Type: 9,
						Pos: token.Pos{
							Filename: "test.tf",
							Offset:   208,
							Line:     12,
							Column:   22,
						},
						Text: "\"list1\"",
						JSON: false,
					},
					{
						Type: 9,
						Pos: token.Pos{
							Filename: "test.tf",
							Offset:   216,
							Line:     12,
							Column:   30,
						},
						Text: "\"list2\"",
						JSON: false,
					},
				},
				Ok: true,
			},
		},
		{
			Name:  "Get another token",
			Input: "literal",
			Result: Result{
				Token: []token.Token{},
				Ok:    false,
			},
		},
		{
			Name:  "Get token by not exists key",
			Input: "unknown",
			Result: Result{
				Token: []token.Token{},
				Ok:    false,
			},
		},
	}

	for _, tc := range cases {
		token, ok := source.GetListToken(tc.Input)
		if !reflect.DeepEqual(token, tc.Result.Token) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(token), pp.Sprint(tc.Result.Token), tc.Name)
		}

		if ok != tc.Result.Ok {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(ok), pp.Sprint(tc.Result.Ok), tc.Name)
		}
	}
}

func TestGetMapToken(t *testing.T) {
	type Result struct {
		Token map[string]token.Token
		Ok    bool
	}

	cases := []struct {
		Name   string
		Input  string
		Result Result
	}{
		{
			Name:  "Get map token",
			Input: "map",
			Result: Result{
				Token: map[string]token.Token{
					"Key": {
						Type: 9,
						Pos: token.Pos{
							Filename: "test.tf",
							Offset:   134,
							Line:     7,
							Column:   12,
						},
						Text: "\"Value\"",
						JSON: false,
					},
				},
				Ok: true,
			},
		},
		{
			Name:  "Get multi map token",
			Input: "multiMap",
			Result: Result{
				Token: map[string]token.Token{
					"Key1": {
						Type: 9,
						Pos: token.Pos{
							Filename: "test.tf",
							Offset:   134,
							Line:     7,
							Column:   12,
						},
						Text: "\"Value1\"",
						JSON: false,
					},
				},
				Ok: true,
			},
		},
		{
			Name:  "Get another token",
			Input: "list",
			Result: Result{
				Token: map[string]token.Token{},
				Ok:    false,
			},
		},
		{
			Name:  "Get token by not exists key",
			Input: "unknown",
			Result: Result{
				Token: map[string]token.Token{},
				Ok:    false,
			},
		},
	}

	for _, tc := range cases {
		token, ok := source.GetMapToken(tc.Input)
		if !reflect.DeepEqual(token, tc.Result.Token) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(token), pp.Sprint(tc.Result.Token), tc.Name)
		}

		if ok != tc.Result.Ok {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(ok), pp.Sprint(tc.Result.Ok), tc.Name)
		}
	}
}

func TestGetAllMapTokens(t *testing.T) {
	type Result struct {
		Token []map[string]token.Token
		Ok    bool
	}

	cases := []struct {
		Name   string
		Input  string
		Result Result
	}{
		{
			Name:  "Get map token",
			Input: "map",
			Result: Result{
				Token: []map[string]token.Token{
					{
						"Key": {
							Type: 9,
							Pos: token.Pos{
								Filename: "test.tf",
								Offset:   134,
								Line:     7,
								Column:   12,
							},
							Text: "\"Value\"",
							JSON: false,
						},
					},
				},
				Ok: true,
			},
		},
		{
			Name:  "Get multi map token",
			Input: "multiMap",
			Result: Result{
				Token: []map[string]token.Token{
					{
						"Key1": {
							Type: 9,
							Pos: token.Pos{
								Filename: "test.tf",
								Offset:   134,
								Line:     7,
								Column:   12,
							},
							Text: "\"Value1\"",
							JSON: false,
						},
					},
					{
						"Key2": token.Token{
							Type: 9,
							Pos: token.Pos{
								Filename: "test2.tf",
								Offset:   134,
								Line:     7,
								Column:   12,
							},
							Text: "\"Value2\"",
							JSON: false,
						},
					},
				},
				Ok: true,
			},
		},
		{
			Name:  "Get another token",
			Input: "list",
			Result: Result{
				Token: []map[string]token.Token{},
				Ok:    false,
			},
		},
		{
			Name:  "Get token by not exists key",
			Input: "unknown",
			Result: Result{
				Token: []map[string]token.Token{},
				Ok:    false,
			},
		},
	}

	for _, tc := range cases {
		token, ok := source.GetAllMapTokens(tc.Input)
		if !reflect.DeepEqual(token, tc.Result.Token) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(token), pp.Sprint(tc.Result.Token), tc.Name)
		}

		if ok != tc.Result.Ok {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(ok), pp.Sprint(tc.Result.Ok), tc.Name)
		}
	}
}
