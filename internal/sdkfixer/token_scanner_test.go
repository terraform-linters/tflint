// Derived from tflint-plugin-sdk v0.25.0/internal/token_scanner_test.go.
// Licensed under MPL-2.0; see LICENSE.
package sdkfixer

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestSeek(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		seek      func(*tokenScanner) error
		wantPos   hcl.Pos
		wantToken hclsyntax.Token
	}{
		{
			name:   "seek to initial position",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				return s.seek(hcl.InitialPos, tokenStart)
			},
			wantPos: hcl.InitialPos,
			wantToken: hclsyntax.Token{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte("foo"),
			},
		},
		{
			name:   "seek to forward with tokenStart",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				return s.seek(hcl.Pos{Line: 1, Column: 7, Byte: 6}, tokenStart)
			},
			wantPos: hcl.Pos{Line: 1, Column: 7, Byte: 6},
			wantToken: hclsyntax.Token{
				Type:  hclsyntax.TokenNumberLit,
				Bytes: []byte("1"),
			},
		},
		{
			name:   "seek to forward with tokenEnd",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				return s.seek(hcl.Pos{Line: 1, Column: 8, Byte: 7}, tokenEnd)
			},
			wantPos: hcl.Pos{Line: 1, Column: 8, Byte: 7},
			wantToken: hclsyntax.Token{
				Type:  hclsyntax.TokenNumberLit,
				Bytes: []byte("1"),
			},
		},
		{
			name:   "seek to backward with tokenStart",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				_ = s.seek(hcl.Pos{Line: 1, Column: 7, Byte: 6}, tokenStart)
				return s.seek(hcl.Pos{Line: 1, Column: 5, Byte: 4}, tokenStart)
			},
			wantPos: hcl.Pos{Line: 1, Column: 5, Byte: 4},
			wantToken: hclsyntax.Token{
				Type:  hclsyntax.TokenEqual,
				Bytes: []byte("="),
			},
		},
		{
			name:   "seek to backward with tokenEnd",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				_ = s.seek(hcl.Pos{Line: 1, Column: 7, Byte: 6}, tokenStart)
				return s.seek(hcl.Pos{Line: 1, Column: 6, Byte: 5}, tokenEnd)
			},
			wantPos: hcl.Pos{Line: 1, Column: 6, Byte: 5},
			wantToken: hclsyntax.Token{
				Type:  hclsyntax.TokenEqual,
				Bytes: []byte("="),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner, diags := newTokenScanner([]byte(test.source), "main.tf")
			if diags.HasErrors() {
				t.Fatalf("failed to set up token scanner: %s", diags)
			}
			if err := test.seek(scanner); err != nil {
				t.Fatalf("failed to seek: %s", err)
			}

			if diff := cmp.Diff(test.wantPos, scanner.pos); diff != "" {
				t.Errorf("position mismatch: %s", diff)
			}
			opt := cmpopts.IgnoreFields(hclsyntax.Token{}, "Range")
			if diff := cmp.Diff(test.wantToken, scanner.token(), opt); diff != "" {
				t.Errorf("token mismatch: %s", diff)
			}
		})
	}
}

func TestScan(t *testing.T) {
	type scanResult struct {
		Pos   hcl.Pos
		Token hclsyntax.Token
	}

	tests := []struct {
		name        string
		source      string
		seek        func(*tokenScanner) error
		scanResults []scanResult
		want        hcl.Pos
	}{
		{
			name:   "scan all tokens",
			source: `foo = 1`,
			seek:   func(s *tokenScanner) error { return nil },
			scanResults: []scanResult{
				{
					Pos: hcl.Pos{Line: 1, Column: 6, Byte: 5},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenEqual,
						Bytes: []byte("="),
					},
				},
				{
					Pos: hcl.Pos{Line: 1, Column: 8, Byte: 7},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenNumberLit,
						Bytes: []byte("1"),
					},
				},
				{
					Pos: hcl.Pos{Line: 1, Column: 8, Byte: 7},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenEOF,
						Bytes: []byte{},
					},
				},
			},
			want: hcl.Pos{Line: 1, Column: 8, Byte: 7},
		},
		{
			name:   "scan tokens from the middle",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				return s.seek(hcl.Pos{Line: 1, Column: 5, Byte: 4}, tokenStart)
			},
			scanResults: []scanResult{
				{
					Pos: hcl.Pos{Line: 1, Column: 8, Byte: 7},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenNumberLit,
						Bytes: []byte("1"),
					},
				},
				{
					Pos: hcl.Pos{Line: 1, Column: 8, Byte: 7},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenEOF,
						Bytes: []byte{},
					},
				},
			},
			want: hcl.Pos{Line: 1, Column: 8, Byte: 7},
		},
		{
			name:   "scan tokens from tokenEnd",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				return s.seek(hcl.Pos{Line: 1, Column: 6, Byte: 5}, tokenEnd)
			},
			scanResults: []scanResult{
				{
					Pos: hcl.Pos{Line: 1, Column: 8, Byte: 7},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenNumberLit,
						Bytes: []byte("1"),
					},
				},
				{
					Pos: hcl.Pos{Line: 1, Column: 8, Byte: 7},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenEOF,
						Bytes: []byte{},
					},
				},
			},
			want: hcl.Pos{Line: 1, Column: 8, Byte: 7},
		},
		{
			name:   "no scan",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				return s.seek(hcl.Pos{Line: 1, Column: 8, Byte: 7}, tokenStart)
			},
			scanResults: []scanResult{},
			want:        hcl.Pos{Line: 1, Column: 8, Byte: 7},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner, diags := newTokenScanner([]byte(test.source), "main.tf")
			if diags.HasErrors() {
				t.Fatalf("failed to set up token scanner: %s", diags)
			}
			if err := test.seek(scanner); err != nil {
				t.Fatalf("failed to seek: %s", err)
			}

			scanResults := []scanResult{}
			for scanner.scan() {
				scanResults = append(scanResults, scanResult{
					Pos:   scanner.pos,
					Token: scanner.token(),
				})
			}

			opt := cmpopts.IgnoreFields(hclsyntax.Token{}, "Range")
			if diff := cmp.Diff(test.scanResults, scanResults, opt); diff != "" {
				t.Errorf("scan result mismatch: %s", diff)
			}
			if diff := cmp.Diff(test.want, scanner.pos); diff != "" {
				t.Errorf("position mismatch: %s", diff)
			}
		})
	}
}

func TestScanBackward(t *testing.T) {
	type scanResult struct {
		Pos   hcl.Pos
		Token hclsyntax.Token
	}

	tests := []struct {
		name        string
		source      string
		seek        func(*tokenScanner) error
		scanResults []scanResult
		want        hcl.Pos
	}{
		{
			name:   "scan all tokens",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				return s.seek(hcl.Pos{Line: 1, Column: 8, Byte: 7}, tokenStart)
			},
			scanResults: []scanResult{
				{
					Pos: hcl.Pos{Line: 1, Column: 7, Byte: 6},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenNumberLit,
						Bytes: []byte("1"),
					},
				},
				{
					Pos: hcl.Pos{Line: 1, Column: 5, Byte: 4},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenEqual,
						Bytes: []byte("="),
					},
				},
				{
					Pos: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte("foo"),
					},
				},
			},
			want: hcl.Pos{Line: 1, Column: 1, Byte: 0},
		},
		{
			name:   "scan tokens from the middle",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				return s.seek(hcl.Pos{Line: 1, Column: 6, Byte: 5}, tokenEnd)
			},
			scanResults: []scanResult{
				{
					Pos: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte("foo"),
					},
				},
			},
			want: hcl.Pos{Line: 1, Column: 1, Byte: 0},
		},
		{
			name:   "scan tokens from tokenStart",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				return s.seek(hcl.Pos{Line: 1, Column: 5, Byte: 4}, tokenStart)
			},
			scanResults: []scanResult{
				{
					Pos: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte("foo"),
					},
				},
			},
			want: hcl.Pos{Line: 1, Column: 1, Byte: 0},
		},
		{
			name:        "no scan",
			source:      `foo = 1`,
			seek:        func(s *tokenScanner) error { return nil },
			scanResults: []scanResult{},
			want:        hcl.Pos{Line: 1, Column: 1, Byte: 0},
		},
		{
			name:   "no scan from endToken",
			source: `foo = 1`,
			seek: func(s *tokenScanner) error {
				s.seekTokenEnd()
				return nil
			},
			scanResults: []scanResult{},
			want:        hcl.Pos{Line: 1, Column: 1, Byte: 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner, diags := newTokenScanner([]byte(test.source), "main.tf")
			if diags.HasErrors() {
				t.Fatalf("failed to set up token scanner: %s", diags)
			}
			if err := test.seek(scanner); err != nil {
				t.Fatalf("failed to seek: %s", err)
			}

			scanResults := []scanResult{}
			for scanner.scanBackward() {
				scanResults = append(scanResults, scanResult{
					Pos:   scanner.pos,
					Token: scanner.token(),
				})
			}

			opt := cmpopts.IgnoreFields(hclsyntax.Token{}, "Range")
			if diff := cmp.Diff(test.scanResults, scanResults, opt); diff != "" {
				t.Errorf("scan result mismatch: %s", diff)
			}
			if diff := cmp.Diff(test.want, scanner.pos); diff != "" {
				t.Errorf("position mismatch: %s", diff)
			}
		})
	}
}
