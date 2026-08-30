package job

import (
	"bytelyon-client/internal/model"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Queue struct {
	s *model.Set[int]
	g sync.WaitGroup
	c chan *model.Bot
	t *time.Ticker
}

func StartQueue() *Queue {
	q := &Queue{
		s: model.NewSet[int](),
		c: make(chan *model.Bot),
		t: time.NewTicker(15 * time.Second),
	}

	for range 3 {
		q.g.Go(func() {
			for b := range q.c {
				Do(b)
				q.s.Delete(b.ID)
			}
		})
	}

	return q
}

func (q *Queue) Poll() {
	log.Log().Msg(`💈 polling`)
	for _, b := range model.GetBots() {
		if q.s.Put(b.ID, true) {
			q.c <- b
		}
	}
}

func (q *Queue) PollChan() <-chan time.Time {
	return q.t.C
}

func (q *Queue) Stop() {
	q.t.Stop()
	close(q.c)
	q.g.Wait()
}
