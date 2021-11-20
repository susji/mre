package match_test

import (
	"fmt"
	"testing"

	"github.com/susji/mre/match"
)

func TestMatchers(t *testing.T) {
	type entry struct {
		matcher match.Node
		desc    string
		left    []rune
		yes, no [][]rune
	}

	table := []entry{
		{
			matcher: match.NewRune('a'),
			desc:    "a",
			yes:     [][]rune{[]rune{'a'}},
			no:      [][]rune{[]rune{'b'}},
		},
		{
			matcher: match.NewNotRune('a'),
			desc:    "!a",
			yes:     [][]rune{[]rune{'b'}},
			no:      [][]rune{[]rune{'a'}},
		},
		{
			matcher: match.NewRuneRange('3', '5'),
			desc:    "3-5",
			yes:     [][]rune{[]rune{'3'}, []rune{'4'}, []rune{'5'}},
			no:      [][]rune{[]rune{'2'}, []rune{}, []rune{'z'}},
		},
		{
			matcher: match.NewOneOrMore(match.NewRune('a')),
			desc:    "a+",
			yes:     [][]rune{[]rune{'a'}, []rune("aaaa")},
			no:      [][]rune{[]rune{'b'}, []rune("ba"), []rune("")},
		},
		{
			matcher: match.NewZeroOrMore(match.NewRune('b')),
			desc:    "b*",
			yes:     [][]rune{[]rune{'b'}, []rune("bbb"), []rune("z")},
			no:      [][]rune{},
		},
		{
			matcher: match.NewAnyOf(
				match.NewRune('a'),
				match.NewRune('b')),
			desc: "a|b",
			yes:  [][]rune{[]rune("a"), []rune("a"), []rune("az")},
			no:   [][]rune{[]rune("z"), []rune("")},
		},
		{
			matcher: match.NewAll(
				match.NewRune('a'),
				match.NewRune('b'),
				match.NewRune('c')),
			desc: "abc",
			yes:  [][]rune{[]rune("abc"), []rune("abcZ")},
			no:   [][]rune{[]rune("abb"), []rune(""), []rune("abC")},
		},
		{
			matcher: match.NewAny(),
			desc:    ".",
			yes:     [][]rune{[]rune("abc"), []rune("abcZ"), []rune("!")},
			no:      [][]rune{[]rune("")},
		},
		{
			matcher: match.NewAll(
				match.NewRune('a'),
				match.NewAnyOf(match.NewRune('b'), match.NewRune('c')),
				match.NewRune('d')),
			desc: "a(b|c)d",
			yes:  [][]rune{[]rune("abd"), []rune("acd")},
			no:   [][]rune{[]rune("aac"), []rune(""), []rune("add")},
		},
		{
			matcher: match.NewN(match.NewRune('a'), 2),
			desc:    "a{2}",
			yes: [][]rune{
				[]rune("aa"), []rune("aab"), []rune("aaa"), []rune("aaaaabbb")},
			no: [][]rune{[]rune("ab"), []rune("a"), []rune("za")},
		},
		{
			matcher: match.NewLengthRange(match.NewRune('a'), 2, 4),
			desc:    "a{2,4}",
			yes: [][]rune{
				[]rune("aa"), []rune("aab"), []rune("aaa"), []rune("aaaaabbb")},
			no: [][]rune{[]rune("ab"), []rune("a"), []rune("za")},
		},
		{
			matcher: match.NewZeroOrOne(match.NewRune('a')),
			desc:    "a?",
			yes: [][]rune{
				[]rune("aa"), []rune("ba"), []rune("z"), []rune(""),
			},
		},
		{
			// NB: our "X times something" matchers are *always* greedy, but
			//     the specific case of ".*?" may be simulated with `ScanTry'.
			matcher: match.NewScanTry(match.NewRune('a')),
			desc:    ".*?a", //
			yes: [][]rune{
				[]rune("ZZZZZa"), []rune("ba"), []rune("a"),
			},
			no: [][]rune{
				[]rune("ZZZZZ"), []rune(""),
			},
		},
		{
			matcher: match.NewExhaustive(match.NewRune('a')),
			desc:    "a$",
			yes:     [][]rune{[]rune("a")},
			no:      [][]rune{[]rune("a "), []rune("aa")},
		},
		{
			matcher: match.NewRune('a'),
			desc:    "^a",
			yes:     [][]rune{[]rune("a"), []rune("aabbccd")},
			no:      [][]rune{[]rune(" a"), []rune(" ")},
		},
	}

	for _, te := range table {
		t.Run(te.desc, func(t *testing.T) {
			for _, y := range te.yes {
				t.Run("should_match_"+string(y), func(t *testing.T) {
					_, _, err := te.matcher.Match(match.NewContext(0), y)
					if err != nil {
						t.Error("did NOT match")
					}
				})
			}
			for _, n := range te.no {
				t.Run("should_NOT_match_"+string(n), func(t *testing.T) {
					_, _, err := te.matcher.Match(match.NewContext(0), n)
					if err == nil {
						t.Error("DID match")
					}
				})
			}
		})
	}
}

func TestMatchersLeft(t *testing.T) {
	type entry struct {
		matcher match.Node
		desc    string
		test    []rune
		left    []rune
	}

	table := []entry{
		{
			matcher: match.NewOneOrMore(match.NewNotRune('a')),
			desc:    "!a+",
			test:    []rune("bba"),
			left:    []rune("a"),
		},
		{
			matcher: match.NewOneOrMore(match.NewRuneRange('1', '2')),
			desc:    "[1-2]+",
			test:    []rune("121222113456"),
			left:    []rune("3456"),
		},
		{
			matcher: match.NewLengthRange(match.NewRune('a'), 2, 4),
			desc:    "a{2,4}",
			test:    []rune("aaac"),
			left:    []rune("c"),
		},
		{
			matcher: match.NewN(match.NewRune('a'), 2),
			desc:    "a{2}",
			test:    []rune("aaac"),
			left:    []rune("ac"),
		},
		{
			matcher: match.NewAll(
				match.NewRune('a'),
				match.NewRune('b'),
				match.NewRune('c')),
			desc: "abc",
			test: []rune("abcdefgh"),
			left: []rune("defgh"),
		},
		{
			matcher: match.NewZeroOrMore(
				match.NewRune('a')),
			desc: "a*",
			test: []rune("aaaaaaaz"),
			left: []rune("z"),
		},
		{
			matcher: match.NewOneOrMore(
				match.NewRune('a')),
			desc: "a*",
			test: []rune("aaaaaaaz"),
			left: []rune("z"),
		},
		{
			matcher: match.NewZeroOrOne(
				match.NewRune('a')),
			desc: "a?",
			test: []rune("aaz"),
			left: []rune("az"),
		},
		{
			matcher: match.NewZeroOrOne(
				match.NewRune('a')),
			desc: "a?",
			test: []rune("z"),
			left: []rune("z"),
		},
	}

	for _, te := range table {
		t.Run(
			fmt.Sprintf(
				"%s with %s should leave %s",
				te.desc,
				string(te.test),
				string(te.left)),
			func(t *testing.T) {
				_, left, err := te.matcher.Match(match.NewContext(0), te.test)
				if err != nil {
					t.Error("should not error")
				}
				if string(left) != string(te.left) {
					t.Errorf(
						"wanted %s but left %s",
						string(te.left),
						string(left))
				}
			})
	}

}
