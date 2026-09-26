package news

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Source string

const (
	BingNews   Source = "Bing News"
	GoogleNews Source = "Google News"
)

func (s Source) URL(query string) string {
	query = strings.ReplaceAll(query, "+", "+")
	query = url.QueryEscape(query)
	switch s {
	case BingNews:
		return fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", query)
	case GoogleNews:
		return fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", query)
	}
	return fmt.Sprintf("Unknown News Source %T", s)
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		XMLName  xml.Name   `xml:"channel"`
		Articles []*Article `xml:"item"`
	} `xml:"channel"`
}

type Article struct {
	XMLName   xml.Name `xml:"item" json:"-"`
	Link      string   `xml:"link" json:"-"`
	Date      string   `xml:"pubDate" json:"published_at"`
	Desc      string   `xml:"description" json:"description"`
	ImgURL    string   `xml:"Image" json:"img_url"`
	Source    string   `xml:"Source" json:"source"`
	Title     string   `xml:"title" json:"title"`
	Body      string   `json:"body"`
	ImgAlt    string   `json:"img_alt"`
	Keywords  []string `json:"keywords"`
	Publisher string   `json:"publisher"`
	URL       string   `json:"url"`
}

func (a *Article) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("t", a.Title).
		Str("#", a.URL).
		Str("@", a.Date).
		Str("d", a.Desc).
		Str("s", a.Source).
		Str("p", a.Publisher).
		Str("i", a.ImgURL).
		Str("a", a.ImgAlt).
		Any("k", a.Keywords).
		Int("b", len(a.Body))
}

func (a *Article) IsGoogleNews() bool {
	return strings.HasPrefix(a.Link, "https://news.google")
}

func (a *Article) PublishedAt() time.Time {
	for _, layout := range []string{time.RFC1123, time.RFC1123Z} {
		if d, err := time.Parse(layout, a.Date); err == nil {
			return d
		}
	}
	return time.Now()
}

func (a *Article) Words() []string {
	return append([]string{
		a.Title,
		a.Body,
		a.Source,
		a.Desc,
		a.ImgAlt,
		a.Publisher,
	}, a.Keywords...)
}
