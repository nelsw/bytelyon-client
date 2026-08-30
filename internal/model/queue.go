package model

import "sync"

// Queue is an unbounded fifo. It buffers the crawl frontier
// so a producer never waits on a consumer that is, in turn,
// waiting on the producer.
type Queue[T any] struct {

	// mutex "locks" the queue for safe multithreaded access.
	mutex sync.Mutex

	// cond handles synchronization amongst asynchronous operations.
	cond sync.Cond

	// items are a generic collection of items.
	items []T

	// closed indicates if the queue is open or closed for bizness.
	closed bool
}

// NewQueue returns a new queue and initializes the necessary variables without the caller needing intimate knowledge.
func NewQueue[T any]() *Queue[T] {
	q := new(Queue[T])
	q.cond.L = &q.mutex
	return q
}

// Push appends an item and wakes a waiting Pop. It never blocks.
func (q *Queue[T]) Push(item T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, item)
	q.cond.Signal()
}

// Pop blocks until an item is available, reporting false once the queue has been closed and drained.
func (q *Queue[T]) Pop() (item T, ok bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// while the queue is empty and not closed
	for len(q.items) == 0 && !q.closed {
		q.cond.Wait()
	}

	// if the queue is closed and empty, return false
	if len(q.items) == 0 {
		return item, false
	}

	// return the first item and remove it from queue items
	item, q.items = q.items[0], q.items[1:]
	return item, true
}

// Close releases every waiting Pop once the remaining items have been drained.
func (q *Queue[T]) Close() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.closed = true
	q.cond.Broadcast()
}
