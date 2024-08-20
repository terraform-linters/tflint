package internal

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// tokenScanner is a token-based scanner for HCL.
// The scanner scans tokens one by one and returns its position and token details.
type tokenScanner struct {
	tokens   hclsyntax.Tokens
	pos      hcl.Pos
	index    int
	filename string
}

func newTokenScanner(source []byte, filename string) (*tokenScanner, hcl.Diagnostics) {
	tokens, diags := hclsyntax.LexConfig(source, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, diags
	}
	return &tokenScanner{
		tokens:   tokens,
		pos:      hcl.InitialPos,
		index:    0,
		filename: filename,
	}, nil
}

type tokenPos int

const (
	tokenStart tokenPos = iota
	tokenEnd
)

// seek moves the currnet position to the given position.
// The destination token is determined by the given match condtion.
//
// match tokenStart:
//
//	  | <- pos
//	foo=1
//	  |-| token is "="
//
// match tokenEnd:
//
//	  | <- pos
//	foo=1
//	|-| token is "foo"
func (s *tokenScanner) seek(to hcl.Pos, match tokenPos) error {
	switch {
	case s.tokenPos(match).Byte == to.Byte:
		return nil
	case to.Byte < s.tokenPos(match).Byte:
		for s.scanBackward() {
			if to.Byte == s.tokenPos(match).Byte {
				s.pos = to
				return nil
			}
		}
	case s.tokenPos(match).Byte < to.Byte:
		for s.scan() {
			if s.tokenPos(match).Byte == to.Byte {
				s.pos = to
				return nil
			}
		}
	default:
		panic("unreachable")
	}

	return fmt.Errorf("no token found at %s:%d,%d", s.filename, to.Line, to.Column)
}

func (s *tokenScanner) seekByIndex(idx int, pos tokenPos) error {
	if idx < 0 || len(s.tokens) <= idx {
		return fmt.Errorf("index out of range: %d", idx)
	}
	s.index = idx
	s.pos = s.tokenPos(pos)
	return nil
}

// seekTokenStart moves the current position to the start of the current token.
func (s *tokenScanner) seekTokenStart() {
	s.pos = s.token().Range.Start
}

func (s *tokenScanner) seekTokenEnd() {
	s.pos = s.token().Range.End
}

// scan moves the current position to the next token.
// position is always set to the end of the token.
func (s *tokenScanner) scan() bool {
	i := s.index + 1
	if i >= len(s.tokens) {
		s.seekTokenEnd()
		return false
	}
	s.index = i
	s.seekTokenEnd()
	return true
}

func (s *tokenScanner) scanIf(tokenType hclsyntax.TokenType) bool {
	i := s.index + 1
	if i >= len(s.tokens) {
		return false
	}
	if s.tokens[i].Type != tokenType {
		return false
	}
	s.scan()
	return true
}

// scanBackward moves the current position to the previous token.
// position is always set to the start of the token.
func (s *tokenScanner) scanBackward() bool {
	i := s.index - 1
	if i < 0 {
		s.seekTokenStart()
		return false
	}
	s.index = i
	s.seekTokenStart()
	return true
}

func (s *tokenScanner) scanBackwardIf(tokenType hclsyntax.TokenType) bool {
	i := s.index - 1
	if i < 0 {
		return false
	}
	if s.tokens[i].Type != tokenType {
		return false
	}
	s.scanBackward()
	return true
}

func (s *tokenScanner) token() hclsyntax.Token {
	return s.tokens[s.index]
}

func (s *tokenScanner) tokenPos(pos tokenPos) hcl.Pos {
	switch pos {
	case tokenStart:
		return s.token().Range.Start
	case tokenEnd:
		return s.token().Range.End
	default:
		panic("unreachable")
	}
}
