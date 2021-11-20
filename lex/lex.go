package lex

import "github.com/susji/mre/token"

func Lex(regexp string) *token.Tokens {
	runes := []rune(regexp)
	toks := &token.Tokens{}

	col := uint(1)

	escaped := false
	for _, r := range runes {
		if r == '\\' && !escaped {
			escaped = true
			col++
			continue
		} else if escaped {
			/*
			 * If we want to support escapes with special meanings, they should
			 * be evaluated here. Otherwise it's a short-circuiting rune-add.
			 */
			toks.Push(token.TOK_RUNE, col, r)
			escaped = false
			continue
		}
		var tk token.TokenKind
		switch r {
		case '^':
			tk = token.TOK_CARET
		case '$':
			tk = token.TOK_DOLLAR
		case '?':
			tk = token.TOK_QU
		case '*':
			tk = token.TOK_STAR
		case '|':
			tk = token.TOK_PIPE
		case '.':
			tk = token.TOK_DOT
		case '{':
			tk = token.TOK_LCURLY
		case '}':
			tk = token.TOK_RCURLY
		case '[':
			tk = token.TOK_LBRACK
		case ']':
			tk = token.TOK_RBRACK
		case '-':
			tk = token.TOK_DASH
		case ',':
			tk = token.TOK_COMMA
		case '+':
			tk = token.TOK_PLUS
		case '(':
			tk = token.TOK_LPAREN
		case ')':
			tk = token.TOK_RPAREN
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			tk = token.TOK_DIGIT
		default:
			tk = token.TOK_RUNE
		}
		toks.Push(tk, col, r)
		col++
	}
	return toks
}
