package model

import (
	"bytelyon-client/internal/provider/api"
	"bytelyon-client/internal/util/url"
	"bytelyon-client/internal/util/uuid"
	"fmt"
)

type Page struct {

	// URL of the page.
	URL string `json:"url"`

	// Domain of the page.
	Domain string `json:"domain"`

	// Title of the page.
	Title string `json:"title"`

	// Meta tags of the page.
	Meta `json:"meta"`

	// Screenshot of the page.
	Screenshot

	// ParentType is the type of page (Serp or Sitemap).
	ParentType string `json:"pageable_type"`

	// ParentID is the ID of the parent entity. This IS required.
	ParentID int `json:"pageable_id"`

	// Index (or rank) of the Search Result as it pertains to its Kind.
	Index int `json:"index"`

	// Kind of Search Result Page (Organic Result, Sponsored Product, etc.)
	Kind string `json:"kind"`
}

func NewPage(b *Bot, URL string, doc *Doc, img []byte, idx int, knd string) *Page {
	return &Page{
		URL:            URL,
		Domain:         url.Domain(URL),
		Title:          doc.Title(),
		Meta:           doc.Meta(),
		ScreenshotData: img,
		ScreenshotKey:  fmt.Sprintf("%s/%d/%s/screenshot.png", b.Type, b.ID, uuid.FromURL(URL)),
		ParentType:     b.Type.Class(),
		ParentID:       b.ChildID,
		Index:          idx,
		Kind:           knd,
	}
}

func (p *Page) Save() {
	var id int
	if p.ParentType == "App\\Models\\Sitemap" {
		id = api.Put(p, "sitemaps", p.ParentID, "pages")
	} else {
		id = api.Put(p, "searches", p.ParentID, "pages")
	}
	p.Screenshot.Save("pages", id)
}
