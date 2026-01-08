package co

import (
	"sync"
	"time"

	syncx "github.com/tempurai/co/internal/syncx"
)

type AsyncBufferTimeSequence[R any, T []R] struct {
	*asyncSequence[T]

	previousIterator Iterator[R]
	interval         time.Duration
}

func NewAsyncBufferTimeSequence[R any, T []R](it AsyncSequenceable[R], interval time.Duration) *AsyncBufferTimeSequence[R, T] {
	a := &AsyncBufferTimeSequence[R, T]{
		previousIterator: it.iterator(),
		interval:         interval,
	}
	a.asyncSequence = NewAsyncSequence[T](a)
	return a
}

func (a *AsyncBufferTimeSequence[R, T]) SetInterval(interval time.Duration) *AsyncBufferTimeSequence[R, T] {
	a.interval = interval
	return a
}

func (c *AsyncBufferTimeSequence[R, T]) iterator() Iterator[T] {
	it := &asyncBufferTimeSequenceIterator[R, T]{
		AsyncBufferTimeSequence: c,
		bufferedData:            NewList[T](),
		bufferWait:              syncx.NewCondx(&sync.Mutex{}),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[T](it)
	return it
}

type asyncBufferTimeSequenceIterator[R any, T []R] struct {
	*asyncSequenceIterator[T]

	*AsyncBufferTimeSequence[R, T]

	previousTime time.Time
	bufferedData *List[T]

	runOnce        sync.Once
	sourceEnded    bool
	bufferWait     *syncx.Condx
	tickerStop     chan struct{}
	tickerStopOnce sync.Once
}

func (it *asyncBufferTimeSequenceIterator[R, T]) intervalPassed() bool {
	return time.Since(it.previousTime) > it.interval
}

func (it *asyncBufferTimeSequenceIterator[R, T]) startBuffer() {
	it.runOnce.Do(func() {
		it.previousTime = time.Now()
		it.tickerStop = make(chan struct{})

		syncx.SafeGo(func() {
			for op := it.previousIterator.next(); op.valid; op = it.previousIterator.next() {
				reachedInterval := it.intervalPassed()
				if it.bufferedData.len() == 0 || reachedInterval {
					it.bufferedData.add(T{})
				}

				lIdx := it.bufferedData.len() - 1
				it.bufferedData.setAt(lIdx, append(it.bufferedData.getAt(lIdx), op.data))

				if reachedInterval {
					it.bufferWait.Broadcastify(&syncx.BroadcastOption{
						PreProcessFn: func() { it.previousTime = time.Now() },
					})
				}
			}
			it.bufferWait.Broadcastify(&syncx.BroadcastOption{
				PreProcessFn: func() {
					it.sourceEnded = true
					it.stopTicker()
				},
			})
		})

		if it.interval > 0 {
			syncx.SafeGo(func() {
				ticker := time.NewTicker(it.interval)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						it.bufferWait.Broadcastify(&syncx.BroadcastOption{})
					case <-it.tickerStop:
						return
					}
				}
			})
		}
	})
}

func (it *asyncBufferTimeSequenceIterator[R, T]) stopTicker() {
	it.tickerStopOnce.Do(func() {
		close(it.tickerStop)
	})
}

func (it *asyncBufferTimeSequenceIterator[R, T]) next() *Optional[T] {
	it.startBuffer()

	it.bufferWait.Waitify(&syncx.WaitOption{
		ConditionFn: func() bool {
			return !it.sourceEnded && (it.bufferedData.len() == 0 || !it.intervalPassed())
		},
	})

	if it.sourceEnded && it.bufferedData.len() == 0 {
		return NewOptionalEmpty[T]()
	}

	return OptionalOf(it.bufferedData.popFirst())
}
