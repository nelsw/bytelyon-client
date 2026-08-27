package main

import (
	"bytelyon-client/internal/bot/news"
	"bytelyon-client/pkg/http"
	"bytelyon-client/pkg/json"
	"bytelyon-client/pkg/url"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

const endpoint = "https://news.google.com/_/DotsSplashUi/data/batchexecute"

var decodeRegex = regexp.MustCompile(`/articles/(?P<encoded_url>[^?]+)`)

type URL string

func (u *URL) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err == nil {
		*u = URL(decodeURL(s))
	}
	return nil
}

type Title string

func (t *Title) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	var s string
	if err = d.DecodeElement(&s, &start); err != nil {
		return
	}

	var ƒ func(*html.Node) bool
	ƒ = func(n *html.Node) (ok bool) {
		if n.Type == html.TextNode {
			*t = Title(s)
			return true
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if ƒ(c) {
				break
			}
		}
		return
	}
	return
}

type Item struct {
	news.Article
	URL       URL    `xml:"link" json:"url"`
	Title     Title  `xml:"description" json:"title"`
	Publisher string `xml:"source" json:"publisher"`
}

func main() {

	q := "btc forecast"
	q = strings.ReplaceAll(q, " ", "+")
	u := fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q)

	rss, err := http.GetXML[news.RSS[Item]](u)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, item := range rss.Channel.Items {
		json.PrettyPrint(item)
	}

}

func decodeURL(gURL string) (URL string) {

	decodeParts := func(signature, timestamp, base64Str string) (result string, err error) {
		body, _ := json.Marshal([][][]any{{{
			"Fbv4je",
			fmt.Sprintf("[\"garturlreq\",[[\"X\",\"X\",[\"X\",\"X\"],null,null,1,1,\"US:en\",null,1,null,null,null,null,null,0,1],\"X\",\"X\",1,[1,1,1],1,1,null,0,0,null,0],\"%s\",%s,\"%s\"]",
				base64Str,
				timestamp,
				signature,
			),
		}}})

		var b []byte
		var payload []any
		var inner []any

		if b, err = http.PostForm(endpoint, url.Values{"f.req": {string(body)}}); err != nil {
			return
		} else if parts := strings.Split(string(b), "\n\n"); len(parts) < 2 {
			return "", errors.New("unexpected batchexecute response format")
		} else if err = json.Unmarshal([]byte(parts[1]), &payload); err != nil {
			return
		} else if len(payload) == 0 {
			return "", errors.New("empty payload")
		} else if entry, ok := payload[0].([]any); !ok || len(entry) < 3 {
			return "", errors.New("unexpected entry structure")
		} else if s, k := entry[2].(string); !k {
			return "", errors.New("missing inner json string")
		} else if err = json.Unmarshal([]byte(s), &inner); err != nil {
			return
		} else if len(inner) < 2 {
			return "", errors.New("unexpected inner array")
		} else if s, k = inner[1].(string); !k {
			return "", errors.New("decoded url not string")
		} else {
			return s, nil
		}
	}

	var decodeNode func(*html.Node, string) (string, error)
	decodeNode = func(n *html.Node, encodedText string) (url string, err error) {
		if n.Type == html.ElementNode && n.Data == "c-wiz" {
			var sg, ts string
			if e := n.FirstChild; e != nil {
				for _, att := range e.Attr {
					if att.Key == "data-n-a-sg" {
						sg = att.Val
					} else if att.Key == "data-n-a-ts" {
						ts = att.Val
					}
				}
			}
			return decodeParts(sg, ts, encodedText)
		}
		// continue traversing every sibling per child. give em noogies.
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if url, err = decodeNode(c, encodedText); url != "" && err == nil {
				break
			}
		}
		return
	}

	l := log.With().Str("gURL", gURL).Logger()

	if matches := decodeRegex.FindStringSubmatch(gURL); len(matches) < 2 {
		l.Warn().Msg("failed to match gstatic regex")
	} else if node, err := http.GetNode(gURL); err != nil {
		l.Warn().Msg("failed to parse gstatic html")
	} else if URL, err = decodeNode(node, matches[1]); err != nil {
		l.Warn().Msg("failed to decode gstatic node")
	} else {
		l.Debug().Str("url", URL).Msg("decoded gstatic url")
	}
	return
}
