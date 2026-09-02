package news

import (
	"encoding/xml"
	"strings"
	"time"

	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/rs/zerolog"
)

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

func (a *Article) IsAfter(t time.Time) bool {
	if d, err := time.Parse(time.RFC1123, a.Date); err != nil {
		return false
	} else {
		return d.After(t)
	}
}

func (a *Article) IsBlacklisted(m map[string]bool) bool {

	if len(m) == 0 {
		return false
	}

	arr := append([]string{
		a.Title,
		a.Body,
		a.Source,
		a.Desc,
		a.ImgAlt,
		a.Publisher,
	}, a.Keywords...)

	for _, sf := range arr {
		for s := range strings.SplitSeq(sf, " ") {
			if _, exists := m[s]; exists {
				return true
			}
		}
	}
	return false
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

func (a *Article) Fill(content string) {

	if a.Publisher = a.Source; strings.HasPrefix(a.Publisher, "https://www.bing") {
		a.Source = "Bing News"
	} else {
		a.Publisher = "Google News"
	}

	if doc, err := model.NewDoc(content); err == nil {
		a.Publisher = doc.Source()
		a.ImgURL = doc.ImgURL(a.ImgURL)
		a.ImgAlt = doc.ImgAlt(a.Title + " - image")
		a.Body = doc.Body()
		a.Keywords = doc.Keywords()
	}

	if a.Publisher == "Bing News" {
		return
	}

	if doc, err := model.NewDoc(a.Desc); err == nil {
		a.Desc = doc.Document.Text()
		a.Source = doc.Find("font").Text()
	}
}
