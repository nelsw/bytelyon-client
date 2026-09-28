package main

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/nelsw/bytelyon-client/internal/testutil"
)

// waitFor polls cond until it holds or the test times out.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	for deadline := time.Now().Add(10 * time.Second); !cond(); time.Sleep(5 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
	}
}

func TestRun(t *testing.T) {
	p := testutil.DB(t)
	m := testutil.Redis(t)
	testutil.Workdir(t)

	// keep a stray SIGINT from killing the test binary before main registers its handler
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	defer signal.Stop(sig)

	done := make(chan struct{})
	go func() {
		defer close(done)
		main()
	}()

	waitFor(t, "subscription", func() bool { return m.PubSubNumSub("bots")["bots"] > 0 })

	m.Publish("bots", "not json")
	m.Publish("bots", `{"ID":0}`)
	m.Publish("bots", `{"id":9,"type":"search","query":"golang","child_id":5}`) // no script, so it fails fast

	waitFor(t, "bot update", func() bool { return len(p.Execs()) == 1 })
	if calls := p.Execs(); calls[0].Args["id"] != 9 {
		t.Errorf("execs = %+v", calls)
	}

	for {
		_ = syscall.Kill(os.Getpid(), syscall.SIGINT)
		select {
		case <-done:
			if !p.Closed {
				t.Error("main should close the database on exit")
			}
			return
		case <-time.After(20 * time.Millisecond):
		}
	}
}
