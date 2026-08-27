package news

import (
	"encoding/xml"
	"fmt"
	"time"
)

type Source string

const (
	BingNews   Source = "Bing News"
	BingSearch Source = "Bing Search"
	GoogleNews Source = "Google News"
)

func (s Source) URL(q string) string {
	switch s {
	case BingNews:
		return fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q)
	case BingSearch:
		return fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q)
	case GoogleNews:
		return fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q)
	}
	return "https://arepublixchickentendersubsonsale.com"
}

// Time wraps time.Time to add custom unmarshaling logic
type Time time.Time

func (t *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	tmp, err := time.Parse(time.RFC1123, s)
	if err != nil {
		return err
	}
	*t = Time(tmp)
	return nil
}

func (t *Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(*t).Format(time.RFC3339) + `"`), nil
}

type RSS[T any] struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		XMLName xml.Name `xml:"channel"`
		Items   []T      `xml:"item"`
	} `xml:"channel"`
}

type Article struct {
	XMLName xml.Name `xml:"item" json:"-"`

	// Title of a news article
	Title string `xml:"title" json:"title"`

	// URL of a news article - not a proxy.
	URL string `xml:"link" json:"url"`

	// PublishedAt timestamp of a news article
	PublishedAt *Time `xml:"pubDate" json:"published_at"`

	// Description of a news article (Bing Search & Bing News)
	Description string `xml:"description" json:"description"`

	// Publisher of a news article (Google News & Bing News)
	Publisher string `xml:"Source" json:"publisher"`

	// ImgURL of a news article thumbnail (Bing News)
	ImgURL string `xml:"Image" json:"img_url"`

	// ImgAlt for the given ImgURL
	ImgAlt string `xml:"-" json:"img_alt"`

	// Source of a news article
	Source Source `xml:"-" json:"source"`

	// Body is (or will be) reader content of a news article.
	Body string `xml:"-" json:"body"`

	// Keywords of a news article (various keyword-related meta tag values)
	Keywords []string `xml:"-" json:"keywords"`
}
