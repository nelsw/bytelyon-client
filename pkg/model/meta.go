package model

import (
	"maps"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Meta map[string][]string

func NewMeta(content string) Meta {
	meta := make(Meta)
	d, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return meta
	}
	d.Find("meta").Each(func(idx int, s *goquery.Selection) {
		k := s.AttrOr("name", s.AttrOr("property", ""))
		v := s.AttrOr("content", "")
		if k, v = strings.TrimSpace(k), strings.TrimSpace(v); k == "" || v == "" {
			return
		}
		m := make(map[string]bool)
		for v = range strings.SplitSeq(v, ",") {
			if v = strings.TrimSpace(v); v != "" {
				m[v] = true
			}
		}
		meta[k] = slices.Sorted(maps.Keys(m))
	})
	return meta
}
