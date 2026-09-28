package search

import (
	"errors"
	"strings"
	"testing"

	"github.com/nelsw/bytelyon-client/internal/testutil"
)

func TestFetch(t *testing.T) {
	p := testutil.DB(t)
	m := testutil.Redis(t)
	testutil.Workdir(t)
	testutil.Script(t, "sync_search", 0)
	testutil.Output(t, "golang", `{"data":{"results":[1,2]}}`)

	sub := m.NewSubscriber()
	sub.Subscribe("evts")
	published := make(chan string, 1)
	go func() { published <- (<-sub.Messages()).Message }() // miniredis blocks publishers until read

	Fetch(1, 4, "golang", true)

	testutil.Eventually(t, "bot update", func() bool { return len(p.Execs()) == 2 })
	if got, want := testutil.Args(t, "sync_search"), "-m true -q golang"; got != want {
		t.Errorf("args = %q, want %q", got, want)
	}

	calls := p.Execs()
	if !strings.Contains(calls[0].SQL, "UPDATE serps") || !strings.Contains(calls[1].SQL, "UPDATE bots") || calls[1].Args["id"] != 1 {
		t.Fatalf("execs = %+v", calls)
	}
	args := calls[0].Args
	if args["id"] != 4 || args["screenshot_key"] != "search/4/golang.png" || args["content_key"] != "search/4/golang.html" {
		t.Errorf("args = %v", args)
	}
	if data := args["data"].(map[string]any); len(data["results"].([]any)) != 2 {
		t.Errorf("data = %v", data)
	}
	if msg := <-published; msg != `{"id": 1, "message": "golang Search result ready!"}` {
		t.Errorf("published %q", msg)
	}
}

func TestFetchFailures(t *testing.T) {
	for name, tt := range map[string]struct {
		code int
		out  string
	}{
		"script fails":   {1, ""},
		"invalid output": {0, "not json"},
	} {
		t.Run(name, func(t *testing.T) {
			p := testutil.DB(t)
			testutil.Workdir(t)
			testutil.Script(t, "sync_search", tt.code)
			testutil.Output(t, "golang", tt.out)

			Fetch(1, 4, "golang", true)

			// the bot is still marked as run, but no serp is saved
			testutil.Eventually(t, "bot update", func() bool { return len(p.Execs()) > 0 })
			if calls := p.Execs(); len(calls) != 1 || !strings.Contains(calls[0].SQL, "UPDATE bots") {
				t.Errorf("only the bot update expected: %+v", calls)
			}
		})
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
