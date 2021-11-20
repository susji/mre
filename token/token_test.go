package token_test

import (
	"testing"

	"github.com/susji/mre/token"
)

func TestTokens(t *testing.T) {
	toks := &token.Tokens{}
	if toks.Count() != 0 {
		t.Fatal("not zero length")
	}
	toks.Push(token.TOK_RUNE, 1, 'a')
	if toks.Count() != 1 {
		t.Fatal("want length of one after push")
	}
	toks.Push(token.TOK_QU, 2, '?')
	if toks.Count() != 2 {
		t.Fatal("want length of two after push")
	}
	if toks.Cur().Kind() != token.TOK_RUNE {
		t.Error("cur should be RUNE")
	}
	if toks.Peek().Kind() != token.TOK_QU {
		t.Error("peek should be QU")
	}

	a := toks.Get()
	if toks.Peek() != nil {
		t.Error("peek after first get should be nil")
	}
	if toks.Count() != 1 {
		t.Error("want length of one after get")
	}
	if toks.Cur().Kind() != token.TOK_QU {
		t.Error("cur should be QU after first get")
	}
	if a.Kind() != token.TOK_RUNE {
		t.Error("wrong kind: ", a.Kind())
	}
	if a.Column() != 1 {
		t.Error("wrong column: ", a.Column())
	}
	if a.Rune() != 'a' {
		t.Error("wrong rune: ", a.Rune())
	}

	b := toks.Get()
	if toks.Cur() != nil {
		t.Error("cur should be nil after two gets")
	}
	if toks.Peek() != nil {
		t.Error("second peek should be nil, too")
	}
	if toks.Count() != 0 {
		t.Fatal("want length of one after two gets")
	}
	if b.Kind() != token.TOK_QU {
		t.Error("wrong kind: ", b.Kind())
	}
	if b.Column() != 2 {
		t.Error("wrong column: ", b.Column())
	}
	if b.Rune() != '?' {
		t.Error("wrong rune: ", b.Rune())
	}

	toks.Push(token.TOK_LPAREN, 3, '(')
	if err := toks.Accept(token.TOK_LPAREN); err != nil {
		t.Errorf("cannot accept after push: %v", err)
	}
	if toks.Count() != 0 {
		t.Errorf("toks should be empty in the end")
	}
}
