package co

import (
	"sync"

	"go.tempura.ink/co/ds/queue"
	syncx "go.tempura.ink/co/internal/sync"
)

type AsyncBufferedChan[R any] struct {
	*asyncSequence[R]

	sourceCh    chan R
	sourceEnded syncx.AtomicBool
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
		syncx.SafeGo(func() {
			for val := range a.sourceCh {
				syncx.CondBroadcast(a.bufferWait, func() {
					a.bufferedData.Enqueue(val)
				})
			}
			syncx.CondBroadcast(a.bufferWait, func() { a.sourceEnded.Set(true) })
		})
	})
}

func (a *AsyncBufferedChan[R]) Complete() *AsyncBufferedChan[R] {
	syncx.SafeClose(a.sourceCh)
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

func (it *asyncBufferedChanIterator[R]) next() *Optional[R] {
	if it.sourceEnded.Get() && it.bufferedData.Len() == 0 {
		return NewOptionalEmpty[R]()
	}
	syncx.CondWait(it.bufferWait, func() bool {
		return !it.sourceEnded.Get() && it.bufferedData.Len() == 0
	})
	if it.sourceEnded.Get() && it.bufferedData.Len() == 0 {
		return NewOptionalEmpty[R]()
	}
	return OptionalOf(it.bufferedData.Dequeue())
}
