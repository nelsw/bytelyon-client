package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mxschmitt/playwright-go"
	"github.com/nelsw/bytelyon-client/internal/news"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"golang.org/x/net/html"
)

const rawHTML = `<a href="https://news.google.com/rss/articles/CBMiWEFVX3lxTE4xdWZ2b3FwTVpqTVdtYm14YTdLTGR0Yy1YOFUxUksxV0RDcGtkMW9yTkp5WnJIV005b0dBVDhPRG5SUExvcHoyanJLX1dhVXdHbV9yRUYySjQ?oc=5" target="_blank">Bitcoin Price Predictions Pit $400K Bulls Against $10K Bears</a>&nbsp;&nbsp;<font color="#6f6f6f">Cryptonews.net</font>`

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	m := map[string]any{
		"url": "https://www.cryptonews.net/bitcoin-price-predictions-pit-400k-bulls-against-10k-bears/",
	}
	out, _ := json.Marshal(m)
	fmt.Println(string(out))
	var a news.Article
	err := json.Unmarshal(out, &a)
	check(err)
	fmt.Println(a)
}

func parseHTML() {
	node, err := html.Parse(strings.NewReader(rawHTML))
	check(err)

	var ƒ func(*html.Node) string
	ƒ = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if t := ƒ(c); t != "" {
				return t
			}
		}
		return ""
	}

	out := ƒ(node)
	fmt.Println(out)
}

func localChrome() {
	runOption := &playwright.RunOptions{
		SkipInstallBrowsers: true,
	}
	err := playwright.Install(runOption)
	if err != nil {
		log.Fatalf("could not install playwright dependencies: %v", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	option := playwright.BrowserTypeLaunchOptions{
		Channel:  new("chrome"),
		Headless: new(false),
	}
	browser, err := pw.Chromium.Launch(option)
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	context, err := play.NewBrowserContext(browser)
	check(err)

	var p playwright.Page
	p, err = play.NewPage(context)
	check(err)

	encodedURL := "https://news.google.com/rss/articles/CBMiWEFVX3lxTE4xdWZ2b3FwTVpqTVdtYm14YTdLTGR0Yy1YOFUxUksxV0RDcGtkMW9yTkp5WnJIV005b0dBVDhPRG5SUExvcHoyanJLX1dhVXdHbV9yRUYySjQ?oc=5"
	_, err = play.GoTo(p, encodedURL)
	check(err)

	err = play.Wait(p, playwright.LoadStateNetworkidle)
	check(err)

	play.Sleep(p, 100000, 250000)

	play.IMG(p, "foo.png")
}
