package co

import (
	"tempura.ink/co/ds/queue"
	co_sync "tempura.ink/co/internal/sync"
)

type asyncFn[R any] func() (R, error)

type AsyncFns[R any] struct {
	*asyncSequence[R]

	fnQueue *queue.Queue[asyncFn[R]]
}

func FromFn[R any]() *AsyncFns[R] {
	a := &AsyncFns[R]{
		fnQueue: queue.NewQueue[asyncFn[R]](),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (c *AsyncFns[R]) AddFns(fns ...func() (R, error)) *AsyncFns[R] {
	for i := range fns {
		c.fnQueue.Enqueue(fns[i])
	}
	return c
}

func (c *AsyncFns[R]) iterator() Iterator[R] {
	it := &asyncFnsIterator[R]{
		AsyncFns: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncFnsIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncFns[R]
}

func (it *asyncFnsIterator[R]) next() *Optional[R] {
	for fn := it.fnQueue.Dequeue(); it.fnQueue.Len() != 0; fn = it.fnQueue.Dequeue() {
		val, err := co_sync.SafeEFn(fn)
		if err != nil {
			it.handleError(err)
			if it.errorMode.shouldSkip() {
				continue
			}
			if it.errorMode.shouldStop() {
				return NewOptionalEmpty[R]()
			}
		}
		return OptionalOf(val)
	}

	return NewOptionalEmpty[R]()
}
