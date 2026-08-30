package model

import (
	"bytelyon-client/internal/provider/http"
	"bytelyon-client/internal/util/url"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

const gStaticEndpoint = "https://news.google.com/_/DotsSplashUi/data/batchexecute"

var decodeRegex = regexp.MustCompile(`/articles/(?P<encoded_url>[^?]+)`)

type GoogleLink string

func (l *GoogleLink) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		*l = GoogleLink(s)
	} else {
		*l = GoogleLink(decodeGoogleLink(s))
	}
	return nil
}

type GoogleTitle string

func (t *GoogleTitle) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	var s string
	if err = d.DecodeElement(&s, &start); err != nil {
		return
	}

	var ƒ func(*html.Node) bool
	ƒ = func(n *html.Node) (ok bool) {
		if n.Type == html.TextNode {
			*t = GoogleTitle(s)
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

type GoogleItem struct {
	XMLName   xml.Name    `xml:"item"`
	Link      GoogleLink  `xml:"link"`
	Title     GoogleTitle `xml:"description"`
	Publisher string      `xml:"source"`
	PubDate   PubDate     `xml:"pubDate"`
}

func decodeGoogleLink(link string) (URL string) {

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

		if b, err = http.PostForm(gStaticEndpoint, url.Values{"f.req": {string(body)}}); err != nil {
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
					switch att.Key {
					case "data-n-a-sg":
						sg = att.Val
					case "data-n-a-ts":
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

	l := log.With().Str("gURL", link).Logger()

	if matches := decodeRegex.FindStringSubmatch(link); len(matches) < 2 {
		l.Warn().Msg("failed to match gstatic regex")
	} else if node, err := http.GetNode(link); err != nil {
		l.Warn().Msg("failed to parse gstatic html")
	} else if URL, err = decodeNode(node, matches[1]); err != nil {
		l.Warn().Msg("failed to decode gstatic node")
	} else {
		log.Debug().Str("url", URL).Msg("decoded gstatic url")
	}
	return
}
