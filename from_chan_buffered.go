package co

import (
	"sync"

	"go.tempura.ink/co/ds/queue"
	syncx "go.tempura.ink/co/internal/syncx"
)

type AsyncBufferedChan[R any] struct {
	*asyncSequence[R]

	sourceCh    chan R
	sourceEnded syncx.AtomicBool
	runOnce     sync.Once

	bufferedData *queue.Queue[R]
	bufferWait   *syncx.Condx
}

func FromChanBuffered[R any](ch chan R) *AsyncBufferedChan[R] {
	a := &AsyncBufferedChan[R]{
		sourceCh:     ch,
		bufferedData: queue.NewQueue[R](),
		bufferWait:   syncx.NewCondx(&sync.Mutex{}),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	a.startListening()
	return a
}

func (a *AsyncBufferedChan[T]) startListening() {
	a.runOnce.Do(func() {
		syncx.SafeGo(func() {
			for val := range a.sourceCh {
				a.bufferWait.Broadcastify(&syncx.BroadcastOption{
					PreProcessFn: func() {
						a.bufferedData.Enqueue(val)
					},
				})
			}
			a.bufferWait.Broadcastify(&syncx.BroadcastOption{
				PreProcessFn: func() { a.sourceEnded.Set(true) },
			})
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
	it.bufferWait.Waitify(&syncx.WaitOption{
		ConditionFn: func() bool {
			return !it.sourceEnded.Get() && it.bufferedData.Len() == 0
		},
	})
	if it.sourceEnded.Get() && it.bufferedData.Len() == 0 {
		return NewOptionalEmpty[R]()
	}
	return OptionalOf(it.bufferedData.Dequeue())
}
