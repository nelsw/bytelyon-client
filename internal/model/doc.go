package model

import (
	"maps"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	textTags = map[string]bool{
		"a":       true,
		"article": true,
		"div":     true,
		"p":       true,
		"h1":      true,
		"h2":      true,
		"h3":      true,
		"h4":      true,
	}
)

type Meta map[string][]string

type Doc struct {
	*goquery.Document
	*html.Node
	body        *string
	description *string
	imgAlt      *string
	imgUrl      *string
	keywords    []string
	meta        Meta
	source      *string
	title       *string
}

func NewDoc(s string) (doc *Doc, err error) {
	doc = new(Doc)
	if doc.Node, err = html.Parse(strings.NewReader(s)); err == nil {
		doc.Document = goquery.NewDocumentFromNode(doc.Node)
	}
	return
}

func (d *Doc) value(keys ...string) string {
	var opts []string
	for _, key := range keys {
		opts = append(opts, d.Meta()[key]...)
	}
	for _, opt := range opts {
		if opt = strings.TrimSpace(opt); opt != "" {
			return opt
		}
	}
	return ""
}

func (d *Doc) Meta() Meta {
	if d.meta != nil {
		return d.meta
	}
	d.meta = make(Meta)
	d.Find("meta").Each(func(idx int, s *goquery.Selection) {
		k := s.AttrOr("name", s.AttrOr("property", ""))
		v := s.AttrOr("content", "")

		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)

		if k == "" || v == "" {
			return
		}

		m := make(map[string]bool)
		for _, v = range strings.Split(v, ",") {
			if v = strings.TrimSpace(v); v != "" {
				m[v] = true
			}
		}
		d.meta[k] = slices.Sorted(maps.Keys(m))
	})
	return d.meta
}

func (d *Doc) MetaValue(key string) []string {
	return d.Meta()[key]
}

func (d *Doc) Title() string {
	if d.title == nil {
		val := d.value("twitter:title", "og:title", "title")
		d.title = &val
	}
	return *d.title
}

func (d *Doc) Keywords() []string {

	if len(d.keywords) > 0 {
		return d.keywords
	}

	var opts []string
	opts = append(opts, d.Meta()["keywords"]...)
	opts = append(opts, d.Meta()["news_keywords"]...)
	opts = append(opts, d.Meta()["article:tag"]...)

	m := make(map[string]bool)
	for _, opt := range opts {
		m[opt] = true
	}

	d.keywords = slices.Sorted(maps.Keys(m))

	return d.keywords
}

func (d *Doc) Source() string {
	if d.source == nil {
		val := d.value("twitter:site", "og:site_name", "og:site")
		d.source = &val
	}
	return *d.source
}

func (d *Doc) Description() string {
	if d.description == nil {
		val := d.value("twitter:description", "og:description", "description", "abstract")
		d.description = &val
	}
	return *d.description
}

func (d *Doc) ImgAlt(def string) string {
	if d.imgAlt == nil {
		val := d.value("twitter:image:alt", "og:image:alt")
		if val == "" {
			val = def
		}
		d.imgAlt = &val
	}
	return *d.imgAlt
}

func (d *Doc) ImgURL(def string) string {
	if d.imgUrl == nil {
		val := d.value("twitter:image:src", "twitter:image", "og:image:secure_url", "og:image:url", "og:image", "image")
		if val == "" {
			val = def
		}
		d.imgUrl = &val
	}
	return *d.imgUrl
}

func (d *Doc) Body() string {

	if d.body != nil {
		return *d.body
	}

	var sel *goquery.Selection
	if sel = d.Find("article"); len(sel.Nodes) == 0 {
		if sel = d.Find("main"); len(sel.Nodes) == 0 {
			if sel = d.Find("body"); len(sel.Nodes) == 0 {
				sel = d.Find("html")
			}
		}
	}

	if sel == nil || len(sel.Nodes) == 0 {
		d.body = new("")
	} else {
		d.body = new(extractText(sel.Nodes[0]))
	}

	return *d.body
}

func (d *Doc) HREFs() []string {
	x := make(map[string]bool)
	d.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			x[href] = true
		}
	})
	return slices.Collect(maps.Keys(x))
}

// extractText recursively wanders through the HTML nodes
func extractText(n *html.Node) string {

	if _, ok := textTags[n.Data]; n.Type == html.ElementNode && !ok {
		return ""
	}

	// If it's a text node, return its data
	if n.Type == html.TextNode {
		return n.Data
	}

	var sb strings.Builder
	// Traverse child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		txt := extractText(c)
		if txt == "" {
			continue
		}
		txt = strings.ReplaceAll(txt, " ,", ",")
		txt = strings.ReplaceAll(txt, " .", ".")
		sb.WriteString(txt)
	}
	return sb.String()
}
