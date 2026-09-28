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

type queue struct {
	ch chan Playable
	wg *sync.WaitGroup
}

var (
	closed bool
	queues []*queue
)

func init() {

	ƒ := func(size int) *queue {
		ch := make(chan Playable)
		var wg sync.WaitGroup
		for range size {
			wg.Go(func() {
				for p := range ch {
					if p.Validate() {
						if out, err := exec.Command(p.Name(), p.Args()...).Output(); err != nil {
							go p.Failure(err)
						} else {
							go p.Success(out)
						}
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

func Close() {
	closed = true
	var wg sync.WaitGroup
	for _, q := range queues {
		wg.Go(func() {
			close(q.ch)
			q.wg.Wait()
		})
	}
	wg.Wait()
}

func It(p Playable) {
	if closed || !p.Validate() {
		return
	}
	if strings.Contains(p.Name(), "/sync_") {
		queues[0].ch <- p
	} else if strings.Contains(p.Name(), "/async_") {
		queues[1].ch <- p
	} else {
		queues[2].ch <- p
	}
}
