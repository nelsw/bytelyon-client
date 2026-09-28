package cache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/nelsw/bytelyon-client/internal/testutil"
	"github.com/nelsw/bytelyon-client/pkg/cache"
	"github.com/redis/go-redis/v9"
)

func TestPutGet(t *testing.T) {
	m := testutil.Redis(t)

	cache.Put(13, "k", "v")
	if got, err := cache.Get(13, "k"); err != nil || got != "v" {
		t.Errorf("Get() = %q, %v", got, err)
	}
	if ttl := m.TTL("k"); ttl != 6*time.Hour {
		t.Errorf("TTL = %v, want 6h", ttl)
	}
	if _, err := cache.Get(13, "missing"); !errors.Is(err, redis.Nil) {
		t.Errorf("Get(missing) error = %v, want redis.Nil", err)
	}
}

func TestUnreachable(t *testing.T) {
	testutil.Isolate()
	cache.Close()
	t.Cleanup(cache.Close)

	cache.Put(13, "k", "v") // logs the ping and set failures
	if _, err := cache.Get(13, "k"); err == nil {
		t.Error("expected error from unreachable redis")
	}
}

func TestSubscribe(t *testing.T) {
	m := testutil.Redis(t)

	got := make(chan string, 2)
	done := make(chan struct{})
	go func() {
		cache.Subscribe(func(payload string) { got <- payload })
		close(done)
	}()

	for len(m.PubSubNumSub("bots")) == 0 || m.PubSubNumSub("bots")["bots"] == 0 {
		time.Sleep(5 * time.Millisecond)
	}
	m.Publish("bots", "one")
	m.Publish("bots", "two")

	for _, want := range []string{"one", "two"} {
		select {
		case p := <-got:
			if p != want {
				t.Errorf("payload = %q, want %q", p, want)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for message")
		}
	}

	cache.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Subscribe did not return after Close")
	}
}
