package mre_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/susji/mre"
)

func TestBasic(t *testing.T) {
	type entry struct {
		expr, desc string
		yes, no    []string
	}

	table := []entry{
		{".", "match any one rune", []string{"z", "yz"}, []string{""}},
		{"^a", "match 'a' rune", []string{"a", "aZ"}, []string{"b", "ba", ""}},
	}

	for _, te := range table {
		t.Run(te.desc, func(t *testing.T) {
			m, err := mre.Compile(te.expr)
			if err != nil {
				t.Error("compile failed: ", err)
			} else if m == nil {
				t.Errorf("compiled without error but nil matcher")
				return
			}

			for _, y := range te.yes {
				t.Run("should_match_"+y, func(t *testing.T) {
					res := m.Match(y)
					if !res {
						t.Errorf("no match")
					}
				})
			}

			for _, n := range te.no {
				t.Run("should_not_match_"+n, func(t *testing.T) {
					res := m.Match(n)
					if res {
						t.Errorf("match")
					}
				})
			}
		})
	}
}

func TestCapture(t *testing.T) {
	m, err := mre.Compile("([12]{2})-([34]{3})")
	if err != nil {
		t.Fatal("compile failed: ", err)
	}
	matched := m.Match("21-434")
	if !matched {
		t.Error("matching failed")
	}
	res := m.Captures()
	exp := []string{
		"21-434",
		"21",
		"434",
	}
	if !reflect.DeepEqual(res, exp) {
		t.Logf("wanted:\n%#v", exp)
		t.Logf("got matches:\n%#v", res)
		t.Fatal("capture mismatch")
	}
}

func TestIpv4(t *testing.T) {
	re := `
^(
	(
		25[0-5]
		|2[0-4][0-9]
		|1[0-9][0-9]
		|[1-9][0-9]
		|[0-9])
	\.)
	{3}

(
	25[0-5]
	|2[0-4][0-9]
	|1[0-9][0-9]
	|[1-9][0-9]
	|[1-9]
)
$`
	for _, rep := range []string{" ", "\t", "\n"} {
		re = strings.ReplaceAll(re, rep, "")
	}

	t.Logf("regex: ``%s''", re)
	m, err := mre.Compile(re)
	if err != nil {
		t.Fatal("compile failed:", err)
	}

	valid := []string{"127.0.0.1", "1.1.1.1", "192.168.1.254"}
	invalid := []string{
		"127.0.0.1.1", "255.255.255.", "256.0.0.1", "0.123.555.1",
		"abc.127.0.1", "127.0.0. 1"}
	for _, v := range valid {
		t.Run("valid_"+v, func(t *testing.T) {
			matched := m.Match(v)
			if !matched {
				t.Fatal("matching failed")
			}
			captures := m.Captures()
			// First three components are in the second capture. Get rid of the
			// last dot manually.
			got := strings.SplitN(captures[1][:], ".", 3)
			got[2] = got[2][:len(got[2])-1]
			got = append(got, captures[3])
			want := strings.Split(v, ".")
			if !reflect.DeepEqual(got, want) {
				t.Errorf("component mismatch, wanted %#v, got %#v",
					want, got)
			}
		})
	}
	for _, i := range invalid {
		t.Run("invalid_"+i, func(t *testing.T) {
			matched := m.Match(i)
			if matched {
				t.Fatal("matching should fail")
			}
		})
	}
}
