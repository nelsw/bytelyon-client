package search

import (
	"errors"
	"strings"
	"testing"

	"github.com/nelsw/bytelyon-client/internal/testutil"
)

func TestFetch(t *testing.T) {
	p := testutil.DB(t)
	testutil.Workdir(t)
	testutil.Script(t, "sync_search", 0)
	testutil.Files(t, ".storage/search/4/golang", `{"results":[1,2]}`)

	Fetch(4, "golang", true)

	if got, want := testutil.Args(t, "sync_search"), "-i 4 -q golang -m true"; got != want {
		t.Errorf("args = %q, want %q", got, want)
	}

	calls := p.Execs()
	if len(calls) != 1 || !strings.Contains(calls[0].SQL, "UPDATE serps") {
		t.Fatalf("execs = %+v", calls)
	}
	args := calls[0].Args
	if args["id"] != 4 || args["screenshot_key"] != "search/4/golang.png" || args["content_key"] != "search/4/golang.html" {
		t.Errorf("args = %v", args)
	}
	if data := args["data"].(map[string]any); len(data["results"].([]any)) != 2 {
		t.Errorf("data = %v", data)
	}
}

func TestFetchScriptFailure(t *testing.T) {
	p := testutil.DB(t)
	testutil.Workdir(t)
	testutil.Script(t, "sync_search", 1)

	Fetch(4, "golang", true)

	if calls := p.Execs(); len(calls) != 0 {
		t.Errorf("nothing should be saved when the script fails: %+v", calls)
	}
}

func TestUpdateSearchError(t *testing.T) {
	p := testutil.DB(t)
	p.ExecErr = errors.New("boom")

	UpdateSearch(1, "img", "src", nil) // logs a warning

	if len(p.Execs()) != 1 {
		t.Error("expected an exec")
	}
}
