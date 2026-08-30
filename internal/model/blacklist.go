package model

import (
	"slices"
	"strings"
)

type Blacklist map[string]bool

func (b *Blacklist) UnmarshalJSON(payload []byte) error {
	m := make(map[string]bool)
	text := strings.ReplaceAll(string(payload), `"`, "")
	for s := range strings.SplitSeq(text, `\n`) {
		if s = strings.TrimSpace(s); s != `null` {
			m[s] = true
		}
	}
	*b = m
	return nil
}

func (b *Blacklist) OK(args ...string) bool {
	for _, arg := range args {
		if slices.ContainsFunc(strings.Split(arg, " "), func(s string) bool { return (*b)[s] }) {
			return false
		}
	}
	return true
}
