package news

import (
	"bytelyon-client/internal/model"
	"bytelyon-client/internal/provider/http"
	"bytelyon-client/internal/util/url"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type BingLink string

func (l *BingLink) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	if v, k := url.Query(s)["url"]; k {
		*l = BingLink(v)
	} else {
		*l = BingLink(s)
	}
	return nil
}

type BingItem struct {
	XMLName xml.Name `xml:"item"`

	// Title of a news article
	Title string `xml:"title"`

	// Link to a news article - not a proxy.
	Link BingLink `xml:"link"`

	// PubDate timestamp of a news article
	PubDate PubDate `xml:"pubDate"`

	// Description of a news article
	Description string `xml:"description"`

	// Source of a news article
	Source string `xml:"Source"`

	// Image GoogleLink of a news article thumbnail
	Image string `xml:"Image"`
}

func DoBing(c chan *model.Article, bot *model.Bot) {

	q := strings.ReplaceAll(bot.Query, " ", "+")
	u := fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q)

	for _, item := range http.GetXML[RSS[BingItem]](u).Channel.Items {
		a := &model.Article{
			Bot:         bot,
			Title:       item.Title,
			URL:         string(item.Link),
			PublishedAt: time.Time(item.PubDate),
			Description: item.Description,
			Publisher:   item.Source,
			ImgURL:      item.Image,
			Source:      "Bing News",
		}

		if a.OK() {
			c <- a
		}
	}
}
