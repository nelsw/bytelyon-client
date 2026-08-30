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

	// ScreenshotKey is the full s3 key for the screenshot.
	ScreenshotKey string `json:"screenshot_key"`

	// ScreenshotData is compressed bytes of a full-page screenshot.
	ScreenshotData []byte `json:"screenshot_data"`

	// Meta tags of the page.
	Meta map[string][]string `json:"meta"`

	// ParentType is the type of page (Serp or Sitemap).
	ParentType string `json:"pageable_type"`

	// ParentID is the ID of the parent entity. This IS required.
	ParentID int `json:"pageable_id"`

	// Index (or rank) of the Search Result as it pertains to its Kind.
	Index int `json:"index"`

	// Kind of Search Result Page (Organic Result, Sponsored Product, etc.)
	Kind string `json:"kind"`
}

func NewSitemapPage(b *Bot, URL string, doc *Doc, img []byte) *Page {
	return &Page{
		URL:            URL,
		Domain:         url.Domain(URL),
		Title:          doc.Title(),
		ScreenshotData: img,
		ScreenshotKey:  fmt.Sprintf("sitemap/%d/%s/screenshot.png", b.ID, uuid.FromURL(URL)),
		ParentType:     "App\\Models\\Sitemap",
		ParentID:       b.SitemapID,
	}
}

func (p *Page) Fill(b *Bot, URL string, doc *Doc, img []byte) {

	// define the primitives
	p.URL = URL
	p.Title = doc.Title()
	p.ScreenshotData = img
	p.Domain = url.Domain(URL)

	// define the parent type, ID
	if b.Type == SearchBot {
		p.ParentID = b.SearchID
		p.ParentType = "App\\Models\\Serp"
	} else {
		p.ParentID = b.SitemapID
		p.ParentType = "App\\Models\\Sitemap"
	}

	// last but not least - the screenshot key
	p.ScreenshotKey = fmt.Sprintf("%s/%s/%d/%s/screenshot.png", b.Type, b.Query, p.ParentID, uuid.FromURL(p.URL))
}

func (p *Page) Save() {
	if p.ParentType == "App\\Models\\Sitemap" {
		api.Put(p, "sitemaps", p.ParentID, "page")
	} else {
		api.Put(p, "searches", p.ParentID, "page")
	}
}
