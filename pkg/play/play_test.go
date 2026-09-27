package play

import (
	"os"
	"testing"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/internal/testutil"
)

func TestScripts(t *testing.T) {
	testutil.Workdir(t)
	testutil.Script(t, "pages", 0)
	testutil.Script(t, "news", 0)
	testutil.Script(t, "sync_search", 0)

	if err := Pages(bot.SitemapType, 7, true, []string{"https://a.com", "https://b.com"}); err != nil {
		t.Fatal(err)
	} else if got, want := testutil.Args(t, "pages"), "-t sitemap -i 7 -m true -u https://a.com https://b.com"; got != want {
		t.Errorf("pages args = %q, want %q", got, want)
	}

	if err := Pages(bot.NewsType, 9, false, []string{"https://n.com"}); err != nil {
		t.Fatal(err)
	} else if got, want := testutil.Args(t, "news"), "-m false -u https://n.com"; got != want {
		t.Errorf("news args = %q, want %q", got, want)
	}

	if err := Search(3, "golang", true); err != nil {
		t.Fatal(err)
	} else if got, want := testutil.Args(t, "sync_search"), "-i 3 -q golang -m true"; got != want {
		t.Errorf("sync_search args = %q, want %q", got, want)
	}
}

func TestScriptFailure(t *testing.T) {
	testutil.Workdir(t)
	testutil.Script(t, "sync_search", 1)

	if err := Search(1, "q", true); err == nil {
		t.Error("expected error from failing script")
	}
	if err := Pages(bot.NewsType, 9, true, nil); err == nil {
		t.Error("expected error from missing script")
	}
}

func TestHandleFiles(t *testing.T) {
	s3 := testutil.Isolate()
	testutil.Workdir(t)
	testutil.Files(t, ".storage/search/1/q", `{"title":"T","n":2}`)

	srcKey, imgKey, data := HandleFiles(".storage/search/1/q")

	if srcKey != "search/1/q.html" || imgKey != "search/1/q.png" {
		t.Errorf("keys = %q, %q", srcKey, imgKey)
	}
	if data.Get("title") != "T" || data.Get("n") != float64(2) {
		t.Errorf("data = %v", data)
	}
	for _, key := range []string{srcKey, imgKey} {
		if _, ok := s3.Object(testutil.Bucket + "/" + key); !ok {
			t.Errorf("%s was not uploaded", key)
		}
	}
	if _, err := os.Stat(".storage/search/1/q.html"); !os.IsNotExist(err) {
		t.Error("html should be removed after upload")
	}
}

func TestHandleFilesMissingOrInvalid(t *testing.T) {
	testutil.Isolate()
	testutil.Workdir(t)

	if _, _, data := HandleFiles(".storage/none"); data == nil || !data.Empty() {
		t.Errorf("missing json should yield empty data, got %v", data)
	}

	testutil.Files(t, ".storage/bad", `{`)
	if _, _, data := HandleFiles(".storage/bad"); data == nil || !data.Empty() {
		t.Errorf("invalid json should yield empty data, got %v", data)
	}
}
