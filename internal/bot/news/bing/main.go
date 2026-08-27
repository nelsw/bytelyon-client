package main

import (
	"bytelyon-client/internal/bot/news"
	"bytelyon-client/pkg/http"
	"bytelyon-client/pkg/json"
	"bytelyon-client/pkg/logs"
	"bytelyon-client/pkg/url"
	"encoding/xml"
	"fmt"
	"strings"
)

type Link string

func (l *Link) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	if v, k := url.Query(s)["url"]; k {
		*l = Link(v)
	} else {
		*l = Link(s)
	}
	return nil
}

func main() {
	logs.Init("debug")
	q := "situation in iran"
	q = strings.ReplaceAll(q, " ", "+")
	u := fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q)

	rss, err := http.GetXML[news.RSS[news.Article]](u)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, item := range rss.Channel.Items {
		json.PrettyPrint(item)
	}

}
