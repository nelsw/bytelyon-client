package news

import (
	"encoding/xml"
	"strings"
	"time"

	"github.com/nelsw/bytelyon-client/pkg/model"

	"github.com/rs/zerolog"
	"golang.org/x/net/html"
)

type Article struct {
	// Title of a news article
	Title string `json:"title"`

	// URL of a news article - not a proxy.
	URL string `json:"url"`

	// PublishedAt timestamp of a news article
	PublishedAt PubDate `json:"published_at" yaml:"published_at"`

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
	evt.Str("t", a.Title).
		Str("#", a.URL).
		Time("@", a.PublishedAt.Time()).
		Str("d", a.Description).
		Str("s", a.Source).
		Str("p", a.Publisher).
		Str("i", a.ImgURL).
		Str("a", a.ImgAlt).
		Any("k", a.Keywords).
		Int("b", len(a.Body))
}

func (a *Article) Fill(content string) {
	doc, err := model.NewDoc(content)
	if err != nil {
		return
	}

	if a.Description == "" {
		a.Description = doc.Description()
	}
	if a.Publisher == "" {
		a.Publisher = doc.Source()
	}
	a.ImgURL = doc.ImgURL(a.ImgURL)
	a.ImgAlt = doc.ImgAlt(a.Title + " - image")
	a.Body = doc.Body()
	a.Keywords = doc.Keywords()
}

type RSS[T any] struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		XMLName xml.Name `xml:"channel"`
		Items   []T      `xml:"item"`
	} `xml:"channel"`
}

// PubDate wraps time.Time to add custom unmarshaling logic
type PubDate time.Time

func (pd *PubDate) After(t time.Time) bool { return pd.Time().After(t) }
func (pd *PubDate) Time() time.Time        { return time.Time(*pd) }

func (pd *PubDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + pd.Time().Format(time.RFC3339) + `"`), nil
}

func (pd *PubDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	tmp, err := time.Parse(time.RFC1123, s)
	if err != nil {
		return err
	}
	*pd = PubDate(tmp)
	return nil
}

type BingItem struct {
	XMLName xml.Name `xml:"item"`

	// Title of a news article
	Title string `xml:"title"`

	// Link to a news article - not a proxy.
	Link string `xml:"link"`

	// PubDate timestamp of a news article
	PubDate PubDate `xml:"pubDate"`

	// Description of a news article
	Description string `xml:"description"`

	// Source of a news article
	Source string `xml:"Source"`

	// Image GoogleLink of a news article thumbnail
	Image string `xml:"Image"`
}

type GoogleTitle string

func (gt *GoogleTitle) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {

	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	node, err := html.Parse(strings.NewReader(s))
	if err != nil {
		*gt = GoogleTitle(s)
		return err
	}

	var ƒ func(*html.Node) string
	ƒ = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if t := ƒ(c); t != s {
				return t
			}
		}
		return s
	}

	*gt = GoogleTitle(ƒ(node))
	return nil
}

func (gt *GoogleTitle) String() string { return string(*gt) }

type GoogleItem struct {
	XMLName   xml.Name    `xml:"item"`
	Link      string      `xml:"link"`
	Title     GoogleTitle `xml:"description"`
	Publisher string      `xml:"source"`
	PubDate   PubDate     `xml:"pubDate"`
}
