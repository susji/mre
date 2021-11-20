package match

import (
	"fmt"
	"strings"
)

const RANGE_UNBOUND = -1

// Interface node describes how a regular expression submatcher should behave.
type Node interface {
	// Match accepts something to parse and returns a possibly
	// non-empty list of matched rune slices, what's left to parse, and a
	// possible error.
	Match(*Context, []rune) ([][]rune, []rune, error)
}

type Context struct {
	ncapturers int
	captures   [][]rune
}

type TimesFunc func(Node) Node

type Root struct {
	n Node
}

type Capture struct {
	n  Node
	id int
}

type Exhaustive struct {
	n Node
}

type ScanTry struct {
	n Node
}

type N struct {
	n Node
	a int
}

type LengthRange struct {
	a, b int
	n    Node
}

type ZeroOrOne struct {
	n Node
}

type OneOrMore struct {
	n Node
}

type ZeroOrMore struct {
	n Node
}

type AnyOf struct {
	n []Node
}

type NotRune struct {
	r rune
}

type All struct {
	n []Node
}

type Any struct {
}

type Rune struct {
	r rune
}

type RuneRange struct {
	a, b rune
}

func (ctx *Context) Reset() {
	ctx.captures = make([][]rune, ctx.ncapturers)
}

func (ctx *Context) Captures() [][]rune {
	return ctx.captures
}

func dump(n Node, b *strings.Builder, level int) {
	ind := "------------------------------------------------------------------"
	in := level*4 + 1
	if in >= len(ind) {
		in = len(ind) - 1
	}
	w := func(s string) {
		b.WriteString(ind[:level*2+1] + s + "\n")
	}
	rec := func(n Node) {
		dump(n, b, level+1)
	}
	switch v := n.(type) {
	case nil:
		w("nil")
	case *ZeroOrOne:
		w("?")
		rec(v.n)
	case *ZeroOrMore:
		w("*")
		rec(v.n)
	case *OneOrMore:
		w("+")
		rec(v.n)
	case *AnyOf:
		w("or")
		for _, oo := range v.n {
			rec(oo)
		}
	case *NotRune:
		w(fmt.Sprintf("! '%c'", v.r))
	case *N:
		w(fmt.Sprintf("{%d}", v.a))
		rec(v.n)
	case *All:
		w("all")
		for _, a := range v.n {
			rec(a)
		}
	case *Any:
		w(".")
	case *LengthRange:
		w(fmt.Sprintf("{%d,%d}", v.a, v.b))
		rec(v.n)
	case *ScanTry:
		w("scantry")
		rec(v.n)
	case *Exhaustive:
		w("exhaustive")
		rec(v.n)
	case *Root:
		w("root")
		rec(v.n)
	case *Capture:
		w(fmt.Sprintf("capture#%d", v.id))
		rec(v.n)
	case *RuneRange:
		w(fmt.Sprintf("'%c'-'%c'", v.a, v.b))
	case *Rune:
		w(fmt.Sprintf("'%c'", v.r))
	default:
		panic(fmt.Sprintf("missing case for %T", n))
	}
}

// Dump builds and returns a textual representation of a matcher tree.
func Dump(n Node) string {
	b := &strings.Builder{}
	dump(n, b, 0)
	return b.String()
}

func (n *Root) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	res, left, err := n.n.Match(ctx, expr)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot match: %w", err)
	}
	// Root node always returns the whole argument if we matched.
	ret := [][]rune{expr}
	return append(ret, res...), left, nil
}

func (n *ZeroOrOne) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	res, left, err := n.n.Match(ctx, expr)
	if err != nil {
		return nil, expr, nil
	} else {
		return res, left, nil
	}
}

func (n *ZeroOrMore) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	total := [][]rune{}
	for {
		res, left, err := n.n.Match(ctx, expr)
		total = append(total, res...)
		if err != nil {
			return total, left, nil
		}
		expr = left
	}
}

func (n *OneOrMore) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	matches := 0
	total := [][]rune{}
	for {
		res, left, err := n.n.Match(ctx, expr)
		total = append(total, res...)
		if err != nil {
			break
		}
		expr = left
		matches++
	}
	if matches == 0 {
		return nil, expr, fmt.Errorf("one-or-more: zero matches")
	}
	return total, expr, nil
}

func (n *N) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	total := [][]rune{}
	left := expr
	for i := 0; i < n.a; i++ {
		var err error
		var res [][]rune
		res, left, err = n.n.Match(ctx, left)
		total = append(total, res...)
		if err != nil {
			return nil, expr, fmt.Errorf("N: not matched")
		}
	}
	return total, left, nil
}

func (n *LengthRange) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	matches := 0
	total := [][]rune{}
	left := expr
	for i := 0; i < n.b; i++ {
		var err error
		var res [][]rune
		res, left, err = n.n.Match(ctx, left)
		total = append(total, res...)
		if err != nil {
			break
		}
		matches++
	}
	if matches < n.a {
		return nil, expr, fmt.Errorf("length-range: not within range")
	}
	return total, left, nil
}

func (n *AnyOf) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	for _, nn := range n.n {
		res, left, err := nn.Match(ctx, expr)
		if err != nil {
			continue
		}
		return res, left, err
	}
	return nil, expr, fmt.Errorf("one-of: nothing matched")
}

func (n *NotRune) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	if len(expr) == 0 {
		return nil, expr, fmt.Errorf("not-rune: empty expr")
	}
	if expr[0] != n.r {
		return [][]rune{[]rune{expr[0]}}, expr[1:], nil
	}
	return nil, expr, fmt.Errorf("not-rune: wanted to avoid %c, but got it", n.r)
}

func (n *Capture) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	if len(expr) == 0 {
		return nil, expr, fmt.Errorf("capture: empty expr")
	}
	res, left, err := n.n.Match(ctx, expr)
	if err == nil && len(res) > 0 {
		for _, sr := range res {
			ctx.captures[n.id] = append(ctx.captures[n.id], sr...)
		}
	}
	return res, left, err
}

func (n *All) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	total := [][]rune{}
	left := expr
	for _, nn := range n.n {
		var err error
		var res [][]rune

		res, left, err = nn.Match(ctx, left)
		if err != nil {
			return nil, expr, fmt.Errorf("all: mismatch")
		}
		total = append(total, res...)
	}
	return total, left, nil
}

func (n *Rune) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	if len(expr) == 0 {
		return nil, expr, fmt.Errorf("rune: empty expr")
	}
	if expr[0] == n.r {
		return [][]rune{[]rune{expr[0]}}, expr[1:], nil
	}
	return nil, expr, fmt.Errorf("rune: wanted %c, got %c", n.r, expr[0])
}

func (n *RuneRange) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	if len(expr) == 0 {
		return nil, expr, fmt.Errorf("rune-range: empty expr")
	}
	if expr[0] >= n.a && expr[0] <= n.b {
		return [][]rune{[]rune{expr[0]}}, expr[1:], nil
	}
	return nil, expr, fmt.Errorf("rune-range: wanted %c-%c, got %c",
		n.a, n.b, expr[0])
}

func (n *Any) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	if len(expr) == 0 {
		return nil, expr, fmt.Errorf("any: empty expr")
	}
	return [][]rune{[]rune{expr[0]}}, expr[1:], nil
}

func (n *Exhaustive) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	if len(expr) == 0 {
		return nil, expr, fmt.Errorf("empty: empty expr")
	}
	res, left, err := n.n.Match(ctx, expr)
	if err != nil {
		return nil, expr, err
	}
	if len(left) > 0 {
		return nil, expr, fmt.Errorf("exhaustive: not exhaustive")
	}
	return res, left, nil
}

func (n *ScanTry) Match(ctx *Context, expr []rune) ([][]rune, []rune, error) {
	cur := expr
	for len(cur) > 0 {
		res, left, err := n.n.Match(ctx, cur)
		if err == nil {
			return res, left, err
		}
		cur = cur[1:]
	}
	return nil, expr, fmt.Errorf("scan-try: no match")
}

func NewScanTry(n Node) Node {
	return &ScanTry{n: n}
}

func NewN(n Node, a int) Node {
	return &N{n: n, a: a}
}

func NewLengthRange(n Node, a, b int) Node {
	return &LengthRange{n: n, a: a, b: b}
}

func NewZeroOrOne(n Node) Node {
	return &ZeroOrOne{n: n}
}

func NewZeroOrMore(n Node) Node {
	return &ZeroOrMore{n: n}
}

func NewOneOrMore(n Node) Node {
	return &OneOrMore{n: n}
}

func NewAnyOf(n ...Node) Node {
	return &AnyOf{n: n}
}

func NewAll(n ...Node) Node {
	return &All{n: n}
}

func NewNotRune(r rune) Node {
	return &NotRune{r: r}
}

func NewRune(r rune) Node {
	return &Rune{r: r}
}

func NewRuneRange(a, b rune) Node {
	return &RuneRange{a: a, b: b}
}

func NewAny() Node {
	return &Any{}
}

func NewExhaustive(n Node) Node {
	return &Exhaustive{n: n}
}

func NewCapture(n Node, id int) Node {
	return &Capture{n: n, id: id}
}

func NewRoot(n Node) Node {
	return &Root{n: n}
}

func NewContext(ncapturers int) *Context {
	return &Context{ncapturers: ncapturers}
}
