package cache_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/nelsw/bytelyon-client/internal/testutil"
	"github.com/nelsw/bytelyon-client/pkg/cache"
	"github.com/redis/go-redis/v9"
)

func TestStrings(t *testing.T) {
	m := testutil.Redis(t)

	cache.SetStr("k", "v")
	if got, err := cache.GetStr("k"); err != nil || got != "v" {
		t.Errorf("GetStr() = %q, %v", got, err)
	}
	if ttl := m.DB(11).TTL("k"); ttl != 6*time.Hour {
		t.Errorf("TTL = %v, want 6h", ttl)
	}
	if _, err := cache.GetStr("missing"); !errors.Is(err, redis.Nil) {
		t.Errorf("GetStr(missing) error = %v, want redis.Nil", err)
	}

	cache.Put(13, "p", "q")
	if got, _ := m.DB(13).Get("p"); got != "q" {
		t.Errorf("Put() stored %q", got)
	}
}

func TestTimes(t *testing.T) {
	testutil.Redis(t)

	now := time.UnixMilli(time.Now().UnixMilli())
	cache.SetTime("t", now)
	if got := cache.GetTime("t"); !got.Equal(now) {
		t.Errorf("GetTime() = %v, want %v", got, now)
	}
	if got := cache.GetTime("missing"); !got.IsZero() {
		t.Errorf("GetTime(missing) = %v, want zero", got)
	}
}

func TestCounters(t *testing.T) {
	testutil.Redis(t)

	for range 2 {
		if err := cache.Incr("n"); err != nil {
			t.Fatal(err)
		}
	}
	if n, err := cache.Decr("n"); err != nil || n != 1 {
		t.Errorf("Decr() = %d, %v", n, err)
	}
	cache.SetStr("n:a", 1)
	cache.SetStr("n:b", 1)
	if keys := cache.Keys("n:*"); !slices.Equal(slices.Sorted(slices.Values(keys)), []string{"n:a", "n:b"}) {
		t.Errorf("Keys() = %v", keys)
	}
	if err := cache.Del("n"); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.GetStr("n"); !errors.Is(err, redis.Nil) {
		t.Errorf("deleted key still present: %v", err)
	}
}

func TestPublish(t *testing.T) {
	m := testutil.Redis(t)
	sub := m.NewSubscriber()
	sub.Subscribe("evts")
	published := make(chan string, 1)
	go func() { published <- (<-sub.Messages()).Message }() // miniredis blocks publishers until read

	cache.Publish(7, "done")

	select {
	case msg := <-published:
		if msg != `{"id": 7, "message": "done"}` {
			t.Errorf("published %q", msg)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestUnreachable(t *testing.T) {
	testutil.Isolate()
	cache.Close()
	t.Cleanup(cache.Close)

	cache.Put(13, "k", "v") // logs the ping and set failures
	if _, err := cache.GetStr("k"); err == nil {
		t.Error("expected error from unreachable redis")
	}
	if keys := cache.Keys("*"); keys != nil {
		t.Errorf("Keys() = %v, want nil", keys)
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
