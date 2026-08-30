package model

import (
	"bytelyon-client/internal/provider/api"
	"time"

	"github.com/rs/zerolog"
)

type Article struct {

	// Bot of a news article
	Bot *Bot `json:"-"`

	// Title of a news article
	Title string `json:"title"`

	// URL of a news article - not a proxy.
	URL string `json:"url"`

	// PublishedAt timestamp of a news article
	PublishedAt time.Time `json:"published_at"`

	// Description of a news article (Bing Search & Bing News)
	Description string `json:"description"`

	// Publisher of a news article (Google News & Bing News)
	Publisher string `json:"publisher"`

	// ImgURL of a news article thumbnail (Bing News)
	ImgURL string `json:"img_url"`

	// ImgAlt for the given ImgURL
	ImgAlt string `json:"img_alt"`

	// Source of a news article
	Source string `json:"source"`

	// Body is browser-reader content of a news article.
	Body string `json:"body"`

	// Keywords of a news article (various keyword-related meta tag values)
	Keywords []string `json:"keywords"`
}

func (a *Article) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("title", a.Title).
		Str("url", a.URL).
		Time("published_at", a.PublishedAt).
		Str("description", a.Description).
		Str("type", a.Source)
}

func (a *Article) Save() {
	api.Put(a, "bots", a.Bot.ID, "articles")
}

func (a *Article) Fill(content string) {
	doc, err := NewDoc(content)
	if err != nil {
		return
	}
	if a.Title == "" {
		a.Title = doc.Title()
	}
	if a.Description == "" {
		a.Description = doc.Description()
	}
	if a.Publisher == "" {
		a.Publisher = doc.Source()
	}
	a.ImgURL = doc.ImgURL(a.ImgURL)
	a.ImgAlt = doc.ImgAlt(a.Bot.Query + " - image")
	a.Body = doc.Body()
	a.Keywords = doc.Keywords()
}

func (a *Article) OK() bool {
	return a.Bot.PlayedAt.Before(a.PublishedAt) && a.Bot.Blacklist.OK(a.Title)
}
