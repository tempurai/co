package co

import (
	"sync"
	"time"

	"github.com/tempura-shrimp/co/ds/queue"
	co_sync "github.com/tempura-shrimp/co/sync"
)

type AsyncDebounceSequence[R any] struct {
	*asyncSequence[R]

	previousIterator Iterator[R]
	interval         time.Duration
	tolerance        time.Duration
}

func NewAsyncDebounceSequence[R any](it AsyncSequenceable[R], interval time.Duration) *AsyncDebounceSequence[R] {
	a := &AsyncDebounceSequence[R]{
		previousIterator: it.iterator(),
		interval:         interval,
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (a *AsyncDebounceSequence[R]) SetInterval(interval time.Duration) *AsyncDebounceSequence[R] {
	a.interval = interval
	return a
}

func (a *AsyncDebounceSequence[R]) SetTolerance(tolerance time.Duration) *AsyncDebounceSequence[R] {
	a.tolerance = a.interval + tolerance
	return a
}

func (c *AsyncDebounceSequence[R]) iterator() Iterator[R] {
	it := &asyncDebounceSequenceIterator[R]{
		AsyncDebounceSequence: c,
		bufferedData:          queue.NewQueue[R](),
		bufferWait:            sync.NewCond(&sync.Mutex{}),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncDebounceSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncDebounceSequence[R]

	previousTime time.Time
	bufferedData *queue.Queue[R]

	runOnce     sync.Once
	sourceEnded bool
	bufferWait  *sync.Cond
}

func (it *asyncDebounceSequenceIterator[R]) intervalPassed() bool {
	return time.Since(it.previousTime) > it.interval
}

func (it *asyncDebounceSequenceIterator[R]) tolerancePassed() bool {
	return time.Since(it.previousTime) > it.tolerance
}

func (it *asyncDebounceSequenceIterator[R]) startBuffer() {
	it.runOnce.Do(func() {
		it.previousTime = time.Now()

		co_sync.SafeGo(func() {
			for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
				if err != nil {
					continue
				}
				reachedInterval := it.intervalPassed()
				if !reachedInterval {
					continue
				}

				it.bufferedData.Enqueue(op.data)
				if it.tolerance == it.interval || it.tolerancePassed() {
					co_sync.CondBoardcast(it.bufferWait, func() { it.previousTime = time.Now() })
				}
			}
			co_sync.CondBoardcast(it.bufferWait, func() { it.sourceEnded = true })
		})
	})
}

func (it *asyncDebounceSequenceIterator[R]) next() (*Optional[R], error) {
	it.startBuffer()

	co_sync.CondWait(it.bufferWait, func() bool {
		return !it.sourceEnded && it.bufferedData.Len() == 0
	})

	if it.sourceEnded && it.bufferedData.Len() == 0 {
		return NewOptionalEmpty[R](), nil
	}

	return OptionalOf(it.bufferedData.Dequeue()), nil
}
