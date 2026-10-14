package classify

import (
	_ "embed"
	"strings"
)

//go:embed stacks.tsv
var stacksTSV string

// Stack is a known TLS client implementation.
type Stack struct {
	Family string
	Name   string
}

var stacks = func() map[string]Stack {
	m := map[string]Stack{}
	for _, line := range strings.Split(stacksTSV, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 3 {
			panic("classify: bad stacks.tsv line: " + line)
		}
		m[f[0]] = Stack{Family: f[1], Name: f[2]}
	}
	return m
}()

// ja4Parts splits a JA4 into what the rules use: the "a" section
// (t13d1516h2: protocol, TLS version, SNI, cipher and extension counts,
// first ALPN) and the cipher hash. ok is false for anything that isn't a
// JA4 (plain HTTP, no hello).
func ja4Parts(ja4 string) (a, cipherHash string, ok bool) {
	a, rest, ok1 := strings.Cut(ja4, "_")
	cipherHash, _, ok2 := strings.Cut(rest, "_")
	if !ok1 || !ok2 || len(a) != 10 {
		return "", "", false
	}
	return a, cipherHash, true
}

// stackFor returns the stack a JA4 belongs to, if it's a known one.
func stackFor(ja4 string) (Stack, bool) {
	_, h, ok := ja4Parts(ja4)
	if !ok {
		return Stack{}, false
	}
	s, ok := stacks[h]
	return s, ok
}
