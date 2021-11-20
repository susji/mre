# MRE

`MRE` is a simple regular expression library. It provides basic capabilities
for matching and extracting string contents. We assume that our input is
UTF-8-encoded.

## Technical details

Our regular expressions support only regular languages, that is, we do not
support backreferences. All matchers are greedy, that is, there is no `?`
suffix.

Subexpressions (`(..)`) imply capturing.

Ranges in set expressions are treated directly with their `uint32` codepoint
values.

If a `regexp` does not begin with `^`, it will be evaluated as containing an
implicit `.*?` in the very beginning. Similarly, if `regexp` does not end with
`$`, it will understood as implicit `.*?` in the very end.

By default, if any of the special characters are to be used for matching
literal runes outside bracketed expressions (sets, they must be escaped with
`\`. Runes within set expressions (`[..]`) are treated literally with the
exception of rune ranges (`-`) and negations (`^`) -- to match them literally,
place them accordingly in bracketed expressions. Otherwise set runes are
matched literally.

XXX Add `]` like POSIX ERE to set matching, ie. for it to be matched as a rune,
it needs to be placed right after `[` or `[^`.

We want alternation (`|`) to bind very loosely and thus we use the traditional
precedence-via-nonterminal-levels approach. The grammar below only deals with
parsing and for this reason escapes are not included as they are treated in the
lexing phase.


Our grammar is roughly the following:

```ebnf
regexp 	= [ "^" ], { or-expr, } [ "$" ]
or-expr = atoms, { "|", atoms }
atoms   = { atom, [ times ] }
atom    = subexpr
        | set
        | "."
        | rune
subexpr = "(", expr, ")"
set     = "[", { "^" }, { rune, [ "-", rune ] }, "]"
times   = "+"
        | "*"
        | "?"
        | "{", posnum, "}"
        | "{", posnum, ",", [ posnum ], "}"
posnum  = "0" | digit, { digit }
digit 	= "0" | ... | "9"
rune 	= any-unicode-codepoint
```
