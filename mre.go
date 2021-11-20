package mre

import (
	"fmt"


	"github.com/susji/mre/compile"
	"github.com/susji/mre/lex"
	"github.com/susji/mre/match"

)

type MRE struct {
	expr string
	root *match.Root
	mctx *match.Context
}

func Compile(expr string) (*MRE, error) {
	toks := lex.Lex(expr)
	if toks.Count() == 0 {
		return nil, fmt.Errorf("Nothing to compile.")
	}

	m := &MRE{}
	mctx, root, err := compile.Compile(toks)
	if err != nil {
		return nil, fmt.Errorf("Compiling failed: %w", err)
	}
	m.mctx = mctx
	m.root = root
	m.expr = expr
	return m, nil
}

func (m *MRE) Match(what string) bool {
	m.mctx.Reset()
	_, _, err := m.root.Match(m.mctx, []rune(what))
	if err != nil {
		return false
	}
	return true
}

func (m *MRE) Dump() string {
	if m.root == nil {
		panic("No matcher.")
	}
	return match.Dump(m.root)
}

func (m *MRE) Captures() []string {
	if m.root == nil {
		panic("No matcher.")
	}
	res := m.mctx.Captures()
	ret := []string{}
	for _, sr := range res {
		ret = append(ret, string(sr))
	}
	return ret
}
