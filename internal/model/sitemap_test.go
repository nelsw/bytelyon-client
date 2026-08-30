package model

import (
	"bytelyon-client/internal/provider/play"
	"fmt"
	"sync"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

func TestParse(t *testing.T) {
	proxy := new(Sitemap)
	for _, tt := range []struct {
		href string
		want string
	}{
		{"/about", "https://bytelyon.com/about"},
		{"about", "https://bytelyon.com/about"},
		{"About/", "https://bytelyon.com/about"},
		{"  /about  ", "https://bytelyon.com/about"},
		{"https://bytelyon.com", "https://bytelyon.com"},
		{"https://bytelyon.com/about", "https://bytelyon.com/about"},
		{"bytelyon.com/about", "https://bytelyon.com/about"},
		{"//bytelyon.com/about", "https://bytelyon.com/about"},
		{"/about#team", "https://bytelyon.com/about"},
		{"/search?q=1", "https://bytelyon.com/search?q=1"},

		{"", ""},
		{"/", ""},
		{"#team", ""},
		{"mailto:hi@bytelyon.com", ""},
		{"javascript:void(0)", ""},
		{"tel:5551234", ""},
		{"http://bytelyon.com/about", ""},
		{"https://example.com/about", ""},
		{"//example.com/about", ""},
		{"/brochure.pdf", ""},
		{"brochure.pdf", ""},
		{"https://bytelyon.com/brochure.pdf", ""},
	} {
		t.Run(tt.href, func(t *testing.T) {
			got, ok := proxy.Parse("bytelyon.com", tt.href)
			if got != tt.want || ok != (tt.want != "") {
				t.Errorf("parse(%q) = %q, %v; want %q, %v", tt.href, got, ok, tt.want, tt.want != "")
			}
		})
	}
}

// TestQueue checks that the frontier hands every item to exactly one consumer and
// releases all of them once closed.
func TestQueue(t *testing.T) {
	const items, consumers = 1000, 8

	q := NewQueue[int]()

	var got sync.Map
	var wg sync.WaitGroup
	for range consumers {
		wg.Go(func() {
			for {
				n, ok := q.Pop()
				if !ok {
					return
				}
				if _, dup := got.LoadOrStore(n, true); dup {
					t.Errorf("item %d popped twice", n)
				}
			}
		})
	}

	for i := range items {
		q.Push(i)
	}
	q.Close()
	wg.Wait()

	for i := range items {
		if _, ok := got.Load(i); !ok {
			t.Fatalf("item %d never popped", i)
		}
	}
}

func TestSitemap(t *testing.T) {

	b := &Bot{Query: "autonation.com"}
	s := NewSitemap(b)
	play.It(false, func(ctx playwright.BrowserContext) {
		s.Build(ctx)
	})
	for _, k := range s.Keys() {
		fmt.Println(k)
	}
}
