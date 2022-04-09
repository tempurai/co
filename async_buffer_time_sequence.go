package co

import (
	"sync"
	"time"

	co_sync "github.com/tempura-shrimp/co/sync"
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
		bufferWait:              sync.NewCond(&sync.Mutex{}),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[T](it)
	return it
}

type asyncBufferTimeSequenceIterator[R any, T []R] struct {
	*asyncSequenceIterator[T]

	*AsyncBufferTimeSequence[R, T]

	previousTime time.Time
	bufferedData *List[T]

	runOnce     sync.Once
	sourceEnded bool
	bufferWait  *sync.Cond
}

func (it *asyncBufferTimeSequenceIterator[R, T]) intervalPassed() bool {
	return time.Since(it.previousTime) > it.interval
}

func (it *asyncBufferTimeSequenceIterator[R, T]) startBuffer() {
	it.runOnce.Do(func() {
		it.previousTime = time.Now()

		co_sync.SafeGo(func() {
			for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
				if err != nil {
					continue
				}

				reachedInterval := it.intervalPassed()
				if it.bufferedData.len() == 0 || reachedInterval {
					it.bufferedData.add(T{})
				}

				lIdx := it.bufferedData.len() - 1
				it.bufferedData.setAt(lIdx, append(it.bufferedData.getAt(lIdx), op.data))

				if reachedInterval {
					co_sync.CondBoardcast(it.bufferWait, func() { it.previousTime = time.Now() })
				}
			}
			co_sync.CondBoardcast(it.bufferWait, func() { it.sourceEnded = true })
		})
	})
}

func (it *asyncBufferTimeSequenceIterator[R, T]) next() (*Optional[T], error) {
	it.startBuffer()

	co_sync.CondWait(it.bufferWait, func() bool {
		return !it.sourceEnded && (it.bufferedData.len() == 0 || !it.intervalPassed())
	})

	if it.sourceEnded && it.bufferedData.len() == 0 {
		return NewOptionalEmpty[T](), nil
	}

	return OptionalOf(it.bufferedData.popFirst()), nil
}
