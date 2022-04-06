package co

import (
	"sync"

	"github.com/tempura-shrimp/co/ds/queue"
	co_sync "github.com/tempura-shrimp/co/sync"
)

type AsyncBufferedChan[R any] struct {
	*asyncSequence[R]

	sourceCh    chan R
	sourceEnded bool
	runOnce     sync.Once

	bufferedData *queue.Queue[R]
	bufferWait   *sync.Cond
}

func FromChanBuffered[R any](ch chan R) *AsyncBufferedChan[R] {
	a := &AsyncBufferedChan[R]{
		sourceCh:     ch,
		bufferedData: queue.NewQueue[R](),
		bufferWait:   sync.NewCond(&sync.Mutex{}),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	a.startListening()
	return a
}

func (a *AsyncBufferedChan[T]) startListening() {
	a.runOnce.Do(func() {
		co_sync.SafeGo(func() {
			for val := range a.sourceCh {
				co_sync.CondBoardcast(a.bufferWait, func() {
					a.bufferedData.Enqueue(val)
				})
			}
			co_sync.CondBoardcast(a.bufferWait, func() { a.sourceEnded = true })
		})
	})
}

func (a *AsyncBufferedChan[R]) Done() *AsyncBufferedChan[R] {
	co_sync.SafeClose(a.sourceCh)
	return a
}

func (a *AsyncBufferedChan[R]) iterator() Iterator[R] {
	it := &asyncBufferedChanIterator[R]{
		AsyncBufferedChan: a,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncBufferedChanIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncBufferedChan[R]
}

func (it *asyncBufferedChanIterator[R]) next() (*Optional[R], error) {
	if it.sourceEnded && it.bufferedData.Len() == 0 {
		return NewOptionalEmpty[R](), nil
	}
	co_sync.CondWait(it.bufferWait, func() bool {
		return !it.sourceEnded && it.bufferedData.Len() == 0
	})
	if it.sourceEnded {
		return NewOptionalEmpty[R](), nil
	}
	return OptionalOf(it.bufferedData.Dequeue()), nil
}
