package news

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/internal/testutil"
	"github.com/rs/zerolog"
)

const (
	googleLink = "https://news.google.com/rss/articles/CBMiABC?oc=5"
	googleURL  = "https://example.com/g"
	bingURL    = "https://example.com/b"
	batchOK    = ")]}'\n\n" + `[["wrb.fr","Fbv4je","[\"garturlres\",\"` + googleURL + `\",1]",null,null,null,"generic"]]`
	articleOK  = `<html><body><c-wiz><div data-n-a-sg="SIG" data-n-a-ts="123" data-other="x"></div></c-wiz></body></html>`
)

// fake serves the Bing and Google endpoints news talks to; a zero status means 200.
type fake struct {
	bing, google, article, batch string
	bingCode, articleCode        int
	batchAbort                   bool
}

func (f *fake) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	write := func(code int, body string) {
		if code != 0 {
			w.WriteHeader(code)
		}
		_, _ = w.Write([]byte(body))
	}
	switch host := r.Header.Get("X-Original-Host"); {
	case host == "www.bing.com":
		write(f.bingCode, f.bing)
	case r.URL.Path == "/rss/search":
		write(0, f.google)
	case strings.HasPrefix(r.URL.Path, "/rss/articles/"):
		write(f.articleCode, f.article)
	case r.URL.Path == "/_/DotsSplashUi/data/batchexecute":
		if f.batchAbort {
			panic(http.ErrAbortHandler)
		}
		write(0, f.batch)
	default:
		http.NotFound(w, r)
	}
}

func rss(items ...string) string {
	return `<rss><channel>` + strings.Join(items, "") + `</channel></rss>`
}

func item(link, title, date, source string) string {
	return fmt.Sprintf(`<item><link>%s</link><title>%s</title><pubDate>%s</pubDate><description>desc</description><Source>%s</Source><Image>https://img</Image></item>`,
		link, title, date, source)
}

func bingLink(u string) string {
	return "https://www.bing.com/news/apiclick.aspx?ref=x&amp;url=" + strings.ReplaceAll(u, ":", "%3a") + "&amp;c=1"
}

func TestSourceURL(t *testing.T) {
	for s, want := range map[Source]string{
		BingNews:   "https://www.bing.com/news/search?format=rss&q=go+lang%26",
		GoogleNews: "https://news.google.com/rss/search?q=go+lang%26&hl=en-US&gl=US&ceid=US:en",
		"other":    "Unknown News Source news.Source",
	} {
		if got := s.URL("go lang&"); got != want {
			t.Errorf("%s.URL() = %q, want %q", s, got, want)
		}
	}
}

func TestArticle(t *testing.T) {
	a := &Article{Link: googleLink, Title: "t", Body: "b", Source: "s", Desc: "d", ImgAlt: "i", Publisher: "p", Keywords: []string{"k"}}
	if !a.IsGoogleNews() || (&Article{Link: "https://bing.com"}).IsGoogleNews() {
		t.Error("IsGoogleNews misclassified links")
	}
	if got := a.Words(); !slices.Equal(got, []string{"t", "b", "s", "d", "i", "p", "k"}) {
		t.Errorf("Words() = %v", got)
	}

	want := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	for _, layout := range []string{time.RFC1123, time.RFC1123Z} {
		if got := (&Article{Date: want.Format(layout)}).PublishedAt(); !got.Equal(want) {
			t.Errorf("PublishedAt(%s) = %v, want %v", layout, got, want)
		}
	}
	if got := (&Article{Date: "garbage"}).PublishedAt(); time.Since(got) > time.Minute {
		t.Errorf("unparseable dates should default to now, got %v", got)
	}

	var buf bytes.Buffer
	l := zerolog.New(&buf)
	l.Log().EmbedObject(&Article{Title: "t", URL: "u", Body: "body"}).Send()
	if out := buf.String(); !strings.Contains(out, `"t":"t"`) || !strings.Contains(out, `"#":"u"`) || !strings.Contains(out, `"b":4`) {
		t.Errorf("MarshalZerologObject output = %s", out)
	}
}

func TestDecodeBingLink(t *testing.T) {
	if got := decodeBingLink("https://bing.com/x?url=https%3a%2f%2fa.com%2fp&y=1"); got != "https://a.com/p" {
		t.Errorf("decodeBingLink() = %q", got)
	}
	if got := decodeBingLink("https://a.com/p"); got != "https://a.com/p" {
		t.Errorf("links without a url param should be returned as-is, got %q", got)
	}
}

func TestDecodeGoogleLink(t *testing.T) {
	m := testutil.Redis(t)
	testutil.Transport(t, &fake{article: articleOK, batch: batchOK})

	if got := decodeGoogleLink(googleLink); got != googleURL {
		t.Fatalf("decodeGoogleLink() = %q, want %q", got, googleURL)
	}
	if cached, _ := m.Get(googleLink); cached != googleURL {
		t.Errorf("decoded url was not cached, got %q", cached)
	}

	// a cache hit skips the network entirely
	_ = m.Set(googleLink, "https://cached.example")
	if got := decodeGoogleLink(googleLink); got != "https://cached.example" {
		t.Errorf("decodeGoogleLink() = %q, want the cached url", got)
	}
}

func TestDecodeGoogleLinkFailures(t *testing.T) {
	tests := map[string]struct {
		link string
		fake fake
	}{
		"no article id":        {"https://news.google.com/rss/topics/x", fake{}},
		"article fetch fails":  {googleLink, fake{articleCode: http.StatusNotFound}},
		"no c-wiz":             {googleLink, fake{article: "<html><body><p>x</p></body></html>"}},
		"post fails":           {googleLink, fake{article: articleOK, batchAbort: true}},
		"single part":          {googleLink, fake{article: articleOK, batch: "only"}},
		"invalid payload":      {googleLink, fake{article: articleOK, batch: "x\n\n{"}},
		"empty payload":        {googleLink, fake{article: articleOK, batch: "x\n\n[]"}},
		"short entry":          {googleLink, fake{article: articleOK, batch: `x` + "\n\n" + `[["a"]]`}},
		"inner not string":     {googleLink, fake{article: articleOK, batch: "x\n\n" + `[["a","b",1]]`}},
		"inner invalid":        {googleLink, fake{article: articleOK, batch: "x\n\n" + `[["a","b","{"]]`}},
		"inner short":          {googleLink, fake{article: articleOK, batch: "x\n\n" + `[["a","b","[1]"]]`}},
		"decoded url not text": {googleLink, fake{article: articleOK, batch: "x\n\n" + `[["a","b","[1,2]"]]`}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			testutil.Redis(t)
			testutil.Transport(t, &tt.fake)
			if got := decodeGoogleLink(tt.link); got != "" {
				t.Errorf("decodeGoogleLink() = %q, want empty", got)
			}
		})
	}
}

// files writes the script output news.Fetch reads for the article at url.
func files(t *testing.T, botID int, url string) {
	t.Helper()
	name := uuid.NewSHA1(uuid.NameSpaceURL, []byte(url)).String()
	testutil.Files(t, filepath.Join(".storage", string(bot.NewsType), strconv.Itoa(botID), name), `{"body":"b"}`)
}

func TestFetch(t *testing.T) {
	p := testutil.DB(t)
	testutil.Redis(t)
	testutil.Workdir(t)
	testutil.Script(t, "pages", 0)

	now := time.Now().UTC()
	fresh, stale := now.Format(time.RFC1123Z), now.Add(-48*time.Hour).Format(time.RFC1123)
	testutil.Transport(t, &fake{
		bing: rss(
			item(bingLink(bingURL), "Bing story", fresh, "Bing Pub"),
			item(bingLink("https://example.com/old"), "Old story", stale, "Bing Pub"),
			item(bingLink("https://example.com/spam"), "spam", fresh, "Bing Pub"),
		),
		google:  rss(item(googleLink, "Google story - Google Pub", fresh, "ignored")),
		article: articleOK,
		batch:   batchOK,
	})
	files(t, 3, bingURL)
	files(t, 3, googleURL)

	var blacklist bot.Blacklist
	_ = blacklist.Scan("spam")

	Fetch(3, "golang", true, now.Add(-time.Hour), blacklist)

	if got, want := testutil.Args(t, "pages"), "-t news -i 3 -m true -u "+bingURL+" "+googleURL; got != want {
		t.Errorf("pages args = %q, want %q", got, want)
	}

	calls := p.Execs()
	if len(calls) != 2 {
		t.Fatalf("expected 2 article upserts, got %+v", calls)
	}
	slices.SortFunc(calls, func(a, b testutil.Call) int { return strings.Compare(a.Args["url"].(string), b.Args["url"].(string)) })

	if b := calls[0].Args; b["url"] != bingURL || b["source"] != "Bing News" || b["publisher"] != "Bing Pub" ||
		b["title"] != "Bing story" || b["description"] != "desc" || b["body"] != "b" || b["bot_id"] != 3 {
		t.Errorf("bing article = %v", b)
	}
	if g := calls[1].Args; g["url"] != googleURL || g["source"] != "Google News" || g["publisher"] != "Google Pub" ||
		g["title"] != "Google story" || g["description"] != "" {
		t.Errorf("google article = %v", g)
	}
}

func TestFetchPartialFailures(t *testing.T) {
	p := testutil.DB(t)
	testutil.Redis(t)
	testutil.Workdir(t)
	testutil.Script(t, "pages", 1)

	testutil.Transport(t, &fake{
		bingCode: http.StatusInternalServerError,
		google:   rss(item("https://news.google.com/rss/topics/x", "Untitled", time.Now().Format(time.RFC1123Z), "")),
	})

	Fetch(3, "golang", false, time.Time{}, bot.Blacklist{})

	// the undecodable google article is still saved (with an empty url) even though the script failed
	if calls := p.Execs(); len(calls) != 1 || calls[0].Args["title"] != "Untitled" {
		t.Errorf("execs = %+v", calls)
	}
}

func TestFetchNothingNew(t *testing.T) {
	p := testutil.DB(t)
	testutil.Workdir(t)
	testutil.Transport(t, &fake{bing: rss(), google: rss()})

	Fetch(3, "golang", false, time.Time{}, bot.Blacklist{})

	if len(p.Execs()) != 0 {
		t.Error("nothing should be saved")
	}
}

func TestUpsertArticleError(t *testing.T) {
	p := testutil.DB(t)
	p.ExecErr = errors.New("boom")

	UpsertArticle(1, &Article{Title: strings.Repeat("x", 300)}) // logs a warning

	if calls := p.Execs(); len(calls) != 1 || len(calls[0].Args["title"].(string)) != 255 {
		t.Errorf("execs = %+v", calls)
	}
}
