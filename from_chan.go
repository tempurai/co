package co

import (
	"sync"

	co_sync "github.com/tempura-shrimp/co/sync"
)

type AsyncChannel[R any] struct {
	*asyncSequence[R]

	sourceCh    chan R
	listenerChs []chan R
	runOnce     sync.Once
}

func FromChan[R any](ch chan R) *AsyncChannel[R] {
	a := &AsyncChannel[R]{
		sourceCh:    ch,
		listenerChs: make([]chan R, 0),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	a.startListening()
	return a
}

func (a *AsyncChannel[T]) startListening() {
	a.runOnce.Do(func() {
		co_sync.SafeGo(func() {
			for val := range a.sourceCh {
				for _, iCh := range a.listenerChs {
					co_sync.SafeSend(iCh, val)
				}
			}
		})
	})
}

func (a *AsyncChannel[R]) Done() *AsyncChannel[R] {
	co_sync.SafeClose(a.sourceCh)
	for i := range a.listenerChs {
		co_sync.SafeClose(a.listenerChs[i])
	}
	return a
}

func (a *AsyncChannel[R]) iterator() Iterator[R] {
	listenerCh := make(chan R)
	a.listenerChs = append(a.listenerChs, listenerCh)

	it := &asyncChannelIterator[R]{
		AsyncChannel: a,
		listenerCh:   listenerCh,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncChannelIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncChannel[R]
	listenerCh chan R
}

func (it *asyncChannelIterator[R]) next() (*Optional[R], error) {
	val, ok := <-it.listenerCh
	if !ok {
		return NewOptionalEmpty[R](), nil
	}

	return OptionalOf(val), nil
}
