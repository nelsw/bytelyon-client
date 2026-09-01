package search

import (
	"fmt"
	"strings"

	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/url"
	"github.com/nelsw/bytelyon-client/pkg/uuid"
)

//go:generate stringer -type=Section
type Section int

const (
	SimilarQueries Section = iota
	SponsoredProducts
	SponsoredResults
	OrganicProducts
	OrganicResults
)

type Result struct {
	url string

	ID int `json:"id"`

	BotID int `json:"bot_id"`

	Query string `json:"query"`

	Ignore map[string]bool `json:"ignore"`

	IMG string `json:"screenshot_key"`

	Data map[Section][]any `json:"data"`
}

func NewResult(id, botID int, query string, ignore []string) *Result {

	r := Result{
		ID:     id,
		BotID:  botID,
		Query:  query,
		Ignore: make(map[string]bool),
		Data:   make(map[Section][]any),
		url:    "https://www.google.com?q=" + strings.ReplaceAll(query, " ", "+"),
	}

	for _, v := range ignore {
		r.Ignore[v] = true
	}

	r.IMG = fmt.Sprintf("search/%d/%s.png", botID, uuid.FromURL(r.url))
	return &r
}

func (r *Result) AddPage(t Section, URL, title, content string) *Page {
	p := &Page{
		Domain: url.Domain(URL),
		Index:  len(r.Data[t]),
		Kind:   t.String(),
		Meta:   model.NewMeta(content),
		IMG:    fmt.Sprintf("search/%d/%s.png", r.BotID, uuid.FromURL(URL)),
		Title:  title,
		URL:    URL,
	}
	r.Add(t, p)
	return p
}

func (r *Result) Add(t Section, a any) {
	if a != nil {
		r.Data[t] = append(r.Data[t], a)
	}
}

type Page struct {
	// URL of the page.
	URL string `json:"url"`

	// Domain of the page.
	Domain string `json:"domain"`

	// Title of the page.
	Title string `json:"title"`

	// Meta tags of the page.
	model.Meta `json:"meta"`

	IMG string `json:"screenshot_key"`

	// Index (or rank) of the Search Result as it pertains to its Kind.
	Index int `json:"index"`

	// Kind of Search Result Page (Organic Result, Sponsored Product, etc.)
	Kind string `json:"kind"`
}
