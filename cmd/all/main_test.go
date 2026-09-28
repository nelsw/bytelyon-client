package main

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/nelsw/bytelyon-client/internal/testutil"
)

// interrupt keeps sending SIGINT to this process until done is closed. The test's own signal.Notify
// ensures an interrupt that arrives before main registers its handler cannot kill the test binary.
func interrupt(t *testing.T, done <-chan struct{}) {
	t.Helper()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	defer signal.Stop(sig)

	deadline := time.After(10 * time.Second)
	for {
		_ = syscall.Kill(os.Getpid(), syscall.SIGINT)
		select {
		case <-done:
			return
		case <-deadline:
			t.Fatal("main did not exit on SIGINT")
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func TestRun(t *testing.T) {
	p := testutil.DB(t)
	testutil.Redis(t)
	testutil.Workdir(t)
	p.Rows = testutil.NewRows([]string{"id", "type", "blacklist", "headless", "query", "last_run_at", "child_id"},
		[]any{1, "search", nil, true, "a", nil, 5}, // no script installed, so the searches fail fast
		[]any{2, "search", nil, true, "b", nil, 6},
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		main()
	}()
	interrupt(t, done)

	if calls := p.Execs(); len(calls) != 2 {
		t.Errorf("expected both due bots to be worked, got %+v", calls)
	}
	if !p.Closed {
		t.Error("main should close the database on exit")
	}
}
