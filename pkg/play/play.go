package play

import (
	"os/exec"
	"strings"
	"sync"
)

type Playable interface {
	Name() string

	Args() []string

	Success([]byte)

	Failure(error)

	Validate() bool
}

// command builds the command that runs p; tests replace it.
var command = func(p Playable) *exec.Cmd { return exec.Command(p.Name(), p.Args()...) }

type queue struct {
	ch chan Playable
	wg *sync.WaitGroup
}

var (
	mu      sync.Mutex
	closed  bool
	pending sync.WaitGroup // jobs accepted by Go or It but not yet queued
	queues  []*queue
)

func init() {
	start()
}

// start launches the worker pools for each queue.
func start() {

	ƒ := func(size int) *queue {
		ch := make(chan Playable)
		var wg sync.WaitGroup
		for range size {
			wg.Go(func() {
				// jobs are validated once, before they're queued, since Validate may count them.
				// results are handled inline so Close waits for them to be saved.
				for p := range ch {
					if out, err := command(p).Output(); err != nil {
						p.Failure(err)
					} else {
						p.Success(out)
					}
				}
			})
		}
		return &queue{ch, &wg}
	}

	queues = []*queue{
		ƒ(5),
		ƒ(20),
		ƒ(50),
	}
}

// Close stops accepting work and waits for accepted jobs to finish; calls after the first are no-ops.
func Close() {
	mu.Lock()
	if closed {
		mu.Unlock()
		return
	}
	closed = true
	mu.Unlock()

	pending.Wait()
	var wg sync.WaitGroup
	for _, q := range queues {
		wg.Go(func() {
			close(q.ch)
			q.wg.Wait()
		})
	}
	wg.Wait()
}

// accept reports whether new work may be queued, registering it as pending if so.
func accept() bool {
	mu.Lock()
	defer mu.Unlock()
	if closed {
		return false
	}
	pending.Add(1)
	return true
}

func enqueue(p Playable) {
	defer pending.Done()
	if strings.Contains(p.Name(), "/sync_") {
		queues[0].ch <- p
	} else if strings.Contains(p.Name(), "/async_") {
		queues[1].ch <- p
	} else {
		queues[2].ch <- p
	}
}

// valid reports whether p should be queued, registering it as pending if so. Validate runs synchronously
// because it may count the job, and callers rely on that count being current once Go returns.
func valid(p Playable) bool {
	if !accept() {
		return false
	} else if !p.Validate() {
		pending.Done()
		return false
	}
	return true
}

// Go queues p in the background; it's dropped if invalid or if Close has been called.
func Go(p Playable) {
	if valid(p) {
		go enqueue(p)
	}
}

// It queues p, blocking until a worker picks it up; it's dropped if invalid or if Close has been called.
func It(p Playable) {
	if valid(p) {
		enqueue(p)
	}
}
