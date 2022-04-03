package co

import (
	"sync"
)

type AsyncPersistentChannel[R any] struct {
	*asyncSequence[R]

	sourceCh    chan R
	sourceEnded bool
	runOnce     sync.Once

	bufferedData *List[R]
	bufferWait   *sync.Cond
}

func OfPersistentChannel[R any](ch chan R) *AsyncPersistentChannel[R] {
	a := &AsyncPersistentChannel[R]{
		sourceCh:     ch,
		bufferedData: NewList[R](),
		bufferWait:   sync.NewCond(&sync.Mutex{}),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	a.startListening()
	return a
}

func (a *AsyncPersistentChannel[T]) startListening() {
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

func (a *AsyncPersistentChannel[R]) Done() *AsyncPersistentChannel[R] {
	SafeClose(a.sourceCh)
	return a
}

func (a *AsyncPersistentChannel[R]) Iterator() Iterator[R] {
	it := &asyncPersistentChannelIterator[R]{
		AsyncPersistentChannel: a,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncPersistentChannelIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncPersistentChannel[R]
	currentIndex int
}

func (it *asyncPersistentChannelIterator[R]) next() (*Optional[R], error) {
	if it.sourceEnded && it.currentIndex >= it.bufferedData.len() {
		return NewOptionalEmpty[R](), nil
	}
	CondWait(it.bufferWait, func() bool {
		return !it.sourceEnded || it.currentIndex >= it.bufferedData.len()
	})

	defer func() { it.currentIndex++ }()
	return OptionalOf(it.bufferedData.getAt(it.currentIndex)), nil
}
