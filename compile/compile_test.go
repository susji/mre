package compile_test

import (
	"reflect"
	"testing"

	"github.com/susji/mre/compile"
	"github.com/susji/mre/lex"
	"github.com/susji/mre/match"
)

func TestCompileBasic(t *testing.T) {
	type entry struct {
		test string
		exp  match.Node
	}

	table := []entry{
		{
			test: "ab+",
			exp: match.NewRoot(
				match.NewScanTry(
					match.NewCapture(
						match.NewAll(
							match.NewRune('a'),
							match.NewOneOrMore(match.NewRune('b'))), 0))),
		},
		{
			test: "^ab+",
			exp: match.NewRoot(
				match.NewCapture(
					match.NewAll(
						match.NewRune('a'),
						match.NewOneOrMore(match.NewRune('b'))), 0)),
		},
		{
			test: "ab+$",
			exp: match.NewRoot(
				match.NewExhaustive(
					match.NewScanTry(
						match.NewCapture(
							match.NewAll(
								match.NewRune('a'),
								match.NewOneOrMore(match.NewRune('b'))), 0)))),
		},
		{
			test: "^ab+$",
			exp: match.NewRoot(
				match.NewExhaustive(
					match.NewCapture(
						match.NewAll(
							match.NewRune('a'),
							match.NewOneOrMore(match.NewRune('b'))), 0))),
		},
		{
			test: "^a?b",
			exp: match.NewRoot(
				match.NewCapture(
					match.NewAll(
						match.NewZeroOrOne(match.NewRune('a')),
						match.NewRune('b')), 0)),
		},
		{
			test: "(ab)+",
			exp: match.NewRoot(
				match.NewScanTry(
					match.NewCapture(
						match.NewOneOrMore(
							match.NewCapture(
								match.NewAll(
									match.NewRune('a'),
									match.NewRune('b')), 1)), 0))),
		},
		{
			test: "((ab){1,3}(c|d))",
			exp: match.NewRoot(
				match.NewScanTry(
					match.NewCapture(
						match.NewCapture(
							match.NewAll(
								match.NewLengthRange(
									match.NewCapture(
										match.NewAll(
											match.NewRune('a'),
											match.NewRune('b')), 2),
									1, 3),
								match.NewCapture(match.NewAnyOf(
									match.NewRune('c'),
									match.NewRune('d')), 3)), 1), 0))),
		},
		{
			test: "^a{1}",
			exp: match.NewRoot(
				match.NewCapture(
					match.NewN(match.NewRune('a'), 1), 0)),
		},
		{
			test: "^a{2,4}",
			exp: match.NewRoot(
				match.NewCapture(
					match.NewLengthRange(match.NewRune('a'), 2, 4), 0)),
		},
		{
			test: "^a{5,}",
			exp: match.NewRoot(match.NewCapture(
				match.NewLengthRange(
					match.NewRune('a'), 5, match.RANGE_UNBOUND), 0)),
		},
		{
			test: "^[-abc\.]+",
			exp: match.NewRoot(
				match.NewCapture(
					match.NewOneOrMore(
						match.NewAnyOf(
							match.NewRune('-'),
							match.NewRune('a'),
							match.NewRune('b'),
							match.NewRune('c'),
							match.NewRune('.'))), 0)),
		},
		{
			test: "^[^ab]",
			exp: match.NewRoot(
				match.NewCapture(
					match.NewAll(
						match.NewNotRune('a'),
						match.NewNotRune('b')), 0)),
		},

		{
			test: "^abc|de?",
			exp: match.NewRoot(
				match.NewCapture(
					match.NewAnyOf(
						match.NewAll(
							match.NewRune('a'),
							match.NewRune('b'),
							match.NewRune('c')),
						match.NewAll(
							match.NewRune('d'),
							match.NewZeroOrOne(match.NewRune('e')))), 0)),
		},
		{
			test: "^a|b|c|de",
			exp: match.NewRoot(
				match.NewCapture(
					match.NewAnyOf(
						match.NewRune('a'),
						match.NewRune('b'),
						match.NewRune('c'),
						match.NewAll(
							match.NewRune('d'),
							match.NewRune('e'))), 0)),
		},
		{
			test: "^[0-9]",
			exp: match.NewRoot(
				match.NewCapture(
					match.NewRuneRange('0', '9'), 0)),
		},
	}

	for _, te := range table {
		t.Run(te.test, func(t *testing.T) {
			_, root, err := compile.Compile(lex.Lex(te.test))
			if err != nil {
				t.Error("errored: ", err)
				return
			}
			if root == nil {
				t.Fatal("root is nil")
			}
			if !reflect.DeepEqual(root, te.exp) {
				t.Error("not equal")
				t.Log("wanted:\n", match.Dump(te.exp))
				t.Log("got:\n", match.Dump(root))
			}
		})
	}
}
