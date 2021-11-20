package lex_test

import (
	"testing"

	"github.com/susji/mre/lex"
	"github.com/susji/mre/token"
)

func TestLexBasic(t *testing.T) {
	type exp struct {
		kind token.TokenKind
		ru   rune
	}
	type entry struct {
		test string
		exp  []exp
	}

	table := []entry{
		entry{
			test: "^(a)+b|cd.[ef-h]{1,39}$",
			exp: []exp{
				exp{token.TOK_CARET, '^'},
				exp{token.TOK_LPAREN, '('},
				exp{token.TOK_RUNE, 'a'},
				exp{token.TOK_RPAREN, ')'},
				exp{token.TOK_PLUS, '+'},
				exp{token.TOK_RUNE, 'b'},
				exp{token.TOK_PIPE, '|'},
				exp{token.TOK_RUNE, 'c'},
				exp{token.TOK_RUNE, 'd'},
				exp{token.TOK_DOT, '.'},
				exp{token.TOK_LBRACK, '['},
				exp{token.TOK_RUNE, 'e'},
				exp{token.TOK_RUNE, 'f'},
				exp{token.TOK_DASH, '-'},
				exp{token.TOK_RUNE, 'h'},
				exp{token.TOK_RBRACK, ']'},
				exp{token.TOK_LCURLY, '{'},
				exp{token.TOK_DIGIT, '1'},
				exp{token.TOK_COMMA, ','},
				exp{token.TOK_DIGIT, '3'},
				exp{token.TOK_DIGIT, '9'},
				exp{token.TOK_RCURLY, '}'},
				exp{token.TOK_DOLLAR, '$'},
			},
		},
		entry{
			test: `\[a\]b\\c`,
			exp: []exp{
				exp{token.TOK_RUNE, '['},
				exp{token.TOK_RUNE, 'a'},
				exp{token.TOK_RUNE, ']'},
				exp{token.TOK_RUNE, 'b'},
				exp{token.TOK_RUNE, '\\'},
				exp{token.TOK_RUNE, 'c'},
			},
		},
	}

	for _, te := range table {
		t.Run(te.test, func(t *testing.T) {
			toks := lex.Lex(te.test)

			if toks.Count() != len(te.exp) {
				t.Errorf("wanted %d tokens, got %d", len(te.exp), toks.Count())
				return
			}

			for i := 0; i < toks.Count(); i++ {
				tok := toks.Get()
				if tok.Kind() != te.exp[i].kind {
					t.Errorf("%d, kind mismatch, wanted '%c'", i, te.exp[i].ru)
				} else if tok.Kind() == token.TOK_RUNE &&
					te.exp[i].ru != tok.Rune() {
					t.Errorf("%d, wanted rune %c, got %c",
						i, te.exp[i].ru, tok.Rune())
				}
			}
		})
	}
}
