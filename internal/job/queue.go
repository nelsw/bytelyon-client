package job

import (
	"bytelyon-client/internal/model"
	"sync"
	"time"
)

type Queue struct {
	x sync.Mutex
	g sync.WaitGroup
	m map[int]bool
	c chan *model.Bot
	t *time.Ticker
}

func NewQueue() *Queue {
	return &Queue{
		m: make(map[int]bool),
		c: make(chan *model.Bot),
		t: time.NewTicker(15 * time.Second),
	}
}

func (q *Queue) Poll() {
	q.x.Lock()
	defer q.x.Unlock()
	for _, b := range model.GetBots() {
		if _, ok := q.m[b.ID]; !ok {
			q.c <- b
		}
	}
}

func (q *Queue) PollInterval() <-chan time.Time {
	return q.t.C
}

func (q *Queue) Start() {

	ƒ := func(b *model.Bot) {

		defer func() {
			q.x.Lock()
			delete(q.m, b.ID)
			q.x.Unlock()
		}()

		Do(b)
	}

	for range 3 {
		for b := range q.c {
			ƒ(b)
		}
	}
}

func (q *Queue) Stop() {
	q.t.Stop()
	close(q.c)
	q.g.Wait()
}
