package compile

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/susji/mre/lex"
	"github.com/susji/mre/match"
	"github.com/susji/mre/token"
)

var bailNestedParens = errors.New("bailing from nested parenthesis")
var bailPipe = errors.New("bailing for or-expr and '|'")
var bailDollar = errors.New("bailing for strict end at '$'")

type ctx struct {
	pardepth   int
	ncapturers int
}

func (ctx *ctx) set(toks *token.Tokens) (match.Node, error) {
	ended, inverse := false, false
	members := []match.Node{}
	// Unless our set is meant to match the inverse ('^'), all our matchers
	// match runes.
	ctor := match.NewRune
	var prevRune rune
	p := func(r rune) {
		members = append(members, ctor(r))
		prevRune = r

	}
	replace := func(r rune, n match.Node) {
		members = members[:len(members)-1]
		members = append(members, n)
		prevRune = r
	}
	fmt.Printf("-> set tokens start: %s\n", toks.Dump())
	// As with POSIX ERE, the set contents are wanted as unmatched literal
	// runes. This means that the position of '-', '^', and ']' matters.
	gotFirstRune := false
	for toks.Count() > 0 && !ended {
		switch toks.Cur().Kind() {
		case token.TOK_CARET:
			// '^' only has non-literal meaning if it is the first rune of the
			// set.
			if !gotFirstRune {
				inverse = true
				ctor = match.NewNotRune
			} else {
				p('^')
			}
		case token.TOK_RBRACK:
			if !gotFirstRune {
				p(']')
			} else {
				ended = true
			}
		case token.TOK_DIGIT:
			gotFirstRune = true
			p(toks.Cur().Rune())
		case token.TOK_RUNE:
			gotFirstRune = true
			p(toks.Cur().Rune())
		case token.TOK_DASH:
			if gotFirstRune {
				// To form a rune range, we need to make sure that we have
				//   - a previous rune
				//   - a next rune.
				//
				// At this point, we need to have at least the dash and the
				// range end runes in our toks.
				if toks.Count() < 2 {
					return nil, fmt.Errorf("set range missing end")
				}
				a := prevRune
				toks.Get()
				b := toks.Cur().Rune()
				if a > b {
					return nil, fmt.Errorf("set range not monotonic")
				}
				// Now, since we already previously pushed the range start to
				// our matchers, we need to pop it off and replace it with the
				// ranged matcher.
				replace(b, match.NewRuneRange(a, b))
			} else {
				p('-')
			}
		case token.TOK_DOT:
			gotFirstRune = true
			p('.')
		}
		toks.Get()
	}
	fmt.Printf("-> set tokens end: %s\n", toks.Dump())
	if !ended {
		return nil, fmt.Errorf("rune set expression missing ']'")
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("rune set expression is empty")
	}
	var b match.Node
	// There are two possibilities for building the set matcher:
	//   - An inverse set, that is, match anything BUT the runes given
	//   - A literal set, that is, match any of the runes
	if len(members) == 1 {
		b = members[0]
	} else if inverse {
		b = match.NewAll(members...)
	} else {
		b = match.NewAnyOf(members...)
	}
	fmt.Printf("-> set expression:\n%s", match.Dump(b))
	return b, nil
}

func (ctx *ctx) atom(toks *token.Tokens) (match.Node, error) {
	tok := toks.Cur()
	fmt.Printf("atom sees %s\n", tok)
	switch tok.Kind() {
	case token.TOK_DOLLAR:
		fmt.Printf("-> atom encountered %s, bailing\n", tok.Name())
		return nil, bailDollar
	case token.TOK_PIPE:
		fmt.Printf("-> atom encountered %s, bailing\n", tok.Name())
		return nil, bailPipe
	case token.TOK_RPAREN:
		fmt.Printf("-> ')' -> pardepth=%d\n", ctx.pardepth)
		if ctx.pardepth < 1 {
			return nil, fmt.Errorf("unbalanced ')'")
		}
		toks.Get()
		ctx.pardepth--
		return nil, bailNestedParens
	case token.TOK_LPAREN:
		toks.Get()
		fmt.Printf("-> '(' -> pardepth=%d\n", ctx.pardepth)
		ctx.pardepth++
		return ctx.orexpr(toks)
	case token.TOK_LBRACK:
		toks.Get()
		fmt.Println("-> '['")
		return ctx.set(toks)
	case token.TOK_DIGIT:
		toks.Get()
		fmt.Printf("-> matched digit '%c'\n", tok.Rune())
		return match.NewRune(tok.Rune()), nil
	case token.TOK_RUNE:
		toks.Get()
		fmt.Printf("-> matched rune '%c'\n", tok.Rune())
		return match.NewRune(tok.Rune()), nil
	case token.TOK_DOT:
		toks.Get()
		return match.NewAny(), nil
	case token.TOK_DASH:
		toks.Get()
		return match.NewRune('-'), nil
	default:
		return nil, fmt.Errorf(
			"unexpected atom: %s [%c]", tok.Name(), tok.Rune())
	}
}

func (ctx *ctx) lengthrange(toks *token.Tokens) (match.TimesFunc, error) {
	tok := toks.Get()
	if tok.Kind() != token.TOK_LCURLY {
		panic("should not happen")
	}
	a := &strings.Builder{}
	earlycurly := false
	gotcomma := false
first_done:
	for toks.Count() != 0 {
		switch toks.Cur().Kind() {
		case token.TOK_DIGIT:
			a.WriteRune(toks.Cur().Rune())
			toks.Get()
		case token.TOK_COMMA:
			toks.Get()
			gotcomma = true
			break first_done
		case token.TOK_RCURLY:
			toks.Get()
			earlycurly = true
			break first_done
		default:
			return nil, fmt.Errorf(
				"erroneous range length definition, got %s",
				toks.Cur().Name())
		}
	}
	if a.Len() == 0 {
		return nil, fmt.Errorf("invalid range length minimum")
	}
	na, _ := strconv.Atoi(a.String())
	// After we parsed the range start, we demand that either
	//   - we do not have the range end at all, ie. it's unbound,
	//     ie. we found '}' before ','
	//   - if we did not '}', then we must have found ','
	//
	// We may end up here also due to tokens running out before either
	// of these conditions was fulfilled.
	if earlycurly {
		fmt.Printf("-> early curly of %d [%s]\n", na, a.String())
		return func(n match.Node) match.Node {
			return match.NewN(
				n,
				na)
		}, nil
	} else if !gotcomma {
		return nil, fmt.Errorf("unterminated length range expression")
	}

	b := &strings.Builder{}
	gotcurly := false
second_done:
	for toks.Count() != 0 {
		switch toks.Cur().Kind() {
		case token.TOK_DIGIT:
			b.WriteRune(toks.Cur().Rune())
			toks.Get()
		case token.TOK_RCURLY:
			toks.Get()
			gotcurly = true
			break second_done
		default:
			return nil, fmt.Errorf(
				"erroneous range length maximum, got %s",
				toks.Cur().Name())
		}
	}
	// Since the upper bound may only end after '}' is received, lack of it
	// means an explicit syntax error.
	if !gotcurly {
		return nil, fmt.Errorf("unterminated range length")
	}
	var nb int
	if b.Len() == 0 {
		nb = match.RANGE_UNBOUND
	} else {
		nb, _ = strconv.Atoi(b.String())
	}
	return func(n match.Node) match.Node {
		return match.NewLengthRange(
			n,
			na, nb)
	}, nil
}

func (ctx *ctx) times(toks *token.Tokens) (match.TimesFunc, error) {
	tok := toks.Cur()
	fmt.Printf("times with %s\n", tok)
	var ret match.TimesFunc
	switch tok.Kind() {
	case token.TOK_PLUS:
		ret = match.NewOneOrMore
	case token.TOK_STAR:
		ret = match.NewZeroOrMore
	case token.TOK_QU:
		ret = match.NewZeroOrOne
	case token.TOK_LCURLY:
		return ctx.lengthrange(toks)
	default:
		fmt.Println("-> no times")
		return nil, nil
	}
	toks.Get()
	return ret, nil
}

func (ctx *ctx) atoms(toks *token.Tokens) (match.Node, error) {
	ret := []match.Node{}
	var reterr error
away:
	for toks.Count() != 0 {
		fmt.Printf("atoms with %s\n", toks.Cur())
		at, err := ctx.atom(toks)
		switch err {
		case nil:
		case bailNestedParens, bailPipe, bailDollar:
			fmt.Println("-> atoms breaking off: ", err)
			reterr = err
			break away
		default:
			return nil, fmt.Errorf("atoms: %w", err)
		}
		if toks.Count() == 0 {
			ret = append(ret, at)
			break
		}
		ti, err := ctx.times(toks)
		if err != nil {
			return nil, err
		} else if ti != nil {
			fmt.Println("-> yes times for atoms")
			ret = append(ret, ti(at))
		} else {
			fmt.Println("-> no times for atoms")
			ret = append(ret, at)
		}
	}
	if len(ret) == 1 {
		return ret[0], reterr
	} else {
		return match.NewAll(ret...), reterr
	}
}

func (ctx *ctx) orexpr(toks *token.Tokens) (match.Node, error) {
	id := ctx.ncapturers
	ctx.ncapturers++
	all := [][]match.Node{[]match.Node{}}
	push := func(n match.Node) {
		ci := len(all) - 1
		all[ci] = append(all[ci], n)
	}
	moreor := func() {
		all = append(all, []match.Node{})
	}
away:
	for toks.Count() != 0 {
		fmt.Printf("orexpr with %s\n", toks.Cur())
		ats, err := ctx.atoms(toks)
		switch err {
		case nil, bailPipe:
			push(ats)
		case bailNestedParens, bailDollar:
			fmt.Println("-> orexpr bailing: ", err)
			push(ats)
			break away
		default:
			return nil, fmt.Errorf("orexpr with atoms failed: %v", err)
		}
		// We rely on the lower level `atom' matcher to kick back here if a '|'
		// is encountered on a suitable place.
		if toks.Count() > 0 && toks.Cur().Kind() == token.TOK_PIPE {
			toks.Get()
			moreor()
		}
	}
	fmt.Printf("-> orexpr gave %d slices\n", len(all))
	// We have two possibilities here:
	//
	//   1) No '|' encountered, ie. a single slice of matchers in all[0], or
	//   2) '|' cases, which means several alternative cases.
	//
	var ret match.Node
	if len(all) == 1 {
		// Optimize the single matcher case again, and avoid redundant `All'
		// wrapping.
		if len(all[0]) == 1 {
			ret = all[0][0]
		} else {
			ret = match.NewAll(all[0]...)
		}
	} else {
		alls := []match.Node{}
		for i := 0; i < len(all); i++ {
			alls = append(alls, all[i]...)
		}
		ret = match.NewAnyOf(alls...)
	}
	return match.NewCapture(ret, id), nil
}

func (ctx *ctx) regexp(toks *token.Tokens) (match.Node, error) {
	if toks.Count() == 0 {
		return nil, fmt.Errorf("no regexp")
	}
	gotCaret, gotDollar := false, false
	if toks.Cur().Kind() == token.TOK_CARET {
		toks.Get()
		gotCaret = true
	}
	re, err := ctx.orexpr(toks)
	if err != nil {
		return nil, err
	}
	if toks.Count() > 0 && toks.Cur().Kind() == token.TOK_DOLLAR {
		toks.Get()
		gotDollar = true
		if toks.Count() > 0 {
			return nil, fmt.Errorf("regexp does not end at '$'")
		}
	}
	var ret match.Node
	switch {
	case !gotCaret && !gotDollar:
		ret = match.NewScanTry(re)
	case gotCaret && gotDollar:
		// ^..$
		ret = match.NewExhaustive(re)
	case gotCaret && !gotDollar:
		// ^..
		ret = re
	default:
		// ..$
		ret = match.NewExhaustive(match.NewScanTry(re))
	}
	return ret, nil
}

func Compile(toks *token.Tokens) (*match.Context, *match.Root, error) {
	ctx := &ctx{ncapturers: 0}

	if toks.Count() == 0 {
		return nil, nil, fmt.Errorf("no tokens to compile")
	}
	re, err := ctx.regexp(toks)
	if err != nil {
		return nil, nil, fmt.Errorf("%v", err)
	}
	mctx := match.NewContext(ctx.ncapturers)
	return mctx, match.NewRoot(re).(*match.Root), nil
}
