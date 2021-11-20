package token

import (
	"fmt"
	"strings"
)

const (
	TOK_STAR = TokenKind(iota)
	TOK_PLUS
	TOK_QU
	TOK_LCURLY
	TOK_RCURLY
	TOK_LBRACK
	TOK_RBRACK
	TOK_LPAREN
	TOK_RPAREN
	TOK_CARET
	TOK_DASH
	TOK_DOLLAR
	TOK_COMMA
	TOK_DOT
	TOK_PIPE
	TOK_DIGIT
	TOK_RUNE
)

var KindNames = []string{
	"*",
	"+",
	"?",
	"{",
	"}",
	"[",
	"]",
	"(",
	")",
	"^",
	"-",
	"$",
	",",
	".",
	"|",
	"D",
	"R",
}

type TokenKind uint8

/* This is a pretty wasteful internal representation. */
type Token struct {
	kind   TokenKind
	column uint
	ru     rune
}

type Tokens struct {
	toks []*Token
}

func (t *Token) Kind() TokenKind {
	return t.kind
}

func (t *Token) Name() string {
	return KindNames[t.kind]
}

func (t *Token) String() string {
	switch t.Kind() {
	case TOK_RUNE:
		return fmt.Sprintf("'%c'", t.Rune())
	case TOK_DIGIT:
		return fmt.Sprintf("<%c>", t.Rune())
	default:
		return t.Name()
	}
}

func (t *Token) Column() uint {
	return t.column
}

func (t *Token) Rune() rune {
	return t.ru
}

func (t *Tokens) Push(kind TokenKind, column uint, ru rune) {
	t.toks = append(t.toks, &Token{kind, column, ru})
}

func (t *Tokens) Count() int {
	return len(t.toks)
}

func (t *Tokens) Get() *Token {
	if len(t.toks) == 0 {
		return nil
	}
	var ret *Token
	ret, t.toks = t.toks[0], t.toks[1:]
	return ret
}

func (t *Tokens) Cur() *Token {
	if len(t.toks) == 0 {
		return nil
	}
	return t.toks[0]
}

func (t *Tokens) Peek() *Token {
	if len(t.toks) < 2 {
		return nil
	}
	return t.toks[1]
}

func (t *Tokens) Accept(tk TokenKind) error {
	if t.Count() < 1 {
		return fmt.Errorf("accept: no more tokens to grab '%s'",
			KindNames[tk])
	}
	if t.Cur().Kind() != tk {
		return fmt.Errorf("accept: expecting '%s' but got '%s'",
			KindNames[tk], t.Cur().Name())
	}
	t.Get()
	return nil
}

func (t *Tokens) Dump() string {
	b := &strings.Builder{}
	for _, t := range t.toks {
		b.WriteString(t.String())
		b.WriteString(",")
	}
	b.WriteString("<END>")
	return b.String()
}
