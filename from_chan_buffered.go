package co

import (
	"sync"
)

type AsyncBufferedChan[R any] struct {
	*asyncSequence[R]

	sourceCh    chan R
	sourceEnded bool
	runOnce     sync.Once

	bufferedData *List[R]
	bufferWait   *sync.Cond
}

func FromChanBuffered[R any](ch chan R) *AsyncBufferedChan[R] {
	a := &AsyncBufferedChan[R]{
		sourceCh:     ch,
		bufferedData: NewList[R](),
		bufferWait:   sync.NewCond(&sync.Mutex{}),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	a.startListening()
	return a
}

func (a *AsyncBufferedChan[T]) startListening() {
	a.runOnce.Do(func() {
		SafeGo(func() {
			for val := range a.sourceCh {
				CondBoardcast(a.bufferWait, func() {
					a.bufferedData.add(val)
				})
			}
			CondBoardcast(a.bufferWait, func() { a.sourceEnded = true })
		})
	})
}

func (a *AsyncBufferedChan[R]) Done() *AsyncBufferedChan[R] {
	SafeClose(a.sourceCh)
	return a
}

func (a *AsyncBufferedChan[R]) Iterator() Iterator[R] {
	it := &asyncBufferedChanIterator[R]{
		AsyncBufferedChan: a,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncBufferedChanIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncBufferedChan[R]
	currentIndex int
}

func (it *asyncBufferedChanIterator[R]) next() (*Optional[R], error) {
	if it.sourceEnded && it.currentIndex >= it.bufferedData.len() {
		return NewOptionalEmpty[R](), nil
	}
	CondWait(it.bufferWait, func() bool {
		return !it.sourceEnded || it.currentIndex >= it.bufferedData.len()
	})

	defer func() { it.currentIndex++ }()
	return OptionalOf(it.bufferedData.getAt(it.currentIndex)), nil
}
