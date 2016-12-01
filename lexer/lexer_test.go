package lexer

import (
	"strings"
	"testing"
)

func newToken(typ TokenType, lit string) Token { return Token{Type: typ, Value: lit} }

func TestScanner(t *testing.T) {
	const input = `5 -5 5.5 + : square ; "string"	for end if else then { 1 } [ 1 ] -`
	scn := NewScanner(strings.NewReader(input))
	if scn == nil {
		t.Fatal("scanner should not be nil")
	}
	var tokens []Token
	for scn.Scan() {
		tokens = append(tokens, scn.Token())
	}
	expected := []Token{
		newToken(Number, "5"),
		newToken(Number, "-5"),
		newToken(Number, "5.5"),
		newToken(Word, "+"),
		newToken(Colon, ":"),
		newToken(Word, "square"),
		newToken(Semicolon, ";"),
		newToken(String, "string"),
		newToken(For, "for"),
		newToken(End, "end"),
		newToken(If, "if"),
		newToken(Else, "else"),
		newToken(Then, "then"),
		newToken(BraceOpen, "{"),
		newToken(Number, "1"),
		newToken(BraceClose, "}"),
		newToken(BracketOpen, "["),
		newToken(Number, "1"),
		newToken(BracketClose, "]"),
		newToken(Word, "-"),
	}
	if len(expected) != len(tokens) {
		t.Fatalf("expecting %d tokens got %d", len(expected), len(tokens))
	}
	for i, v := range tokens {
		if v.Type != expected[i].Type {
			t.Errorf("%d. expecting type %d got %d with value: %s", i, expected[i].Type, v.Type, v.Value)
		}
		if v.Value != expected[i].Value {
			t.Errorf("%d. expecting value %s got %s", i, expected[i].Value, v.Value)
		}
	}
}

func TestLexUnterminatedString(t *testing.T) {
	const input = `"unterminated string`
	scn := NewScanner(strings.NewReader(input))
	for scn.Scan() {
		scn.Token()
	}
	if scn.Err() != nil {
		t.Errorf("expecting nil error, got %v", scn.Err())
	}
}
