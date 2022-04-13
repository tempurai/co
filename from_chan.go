package co

import (
	co_sync "github.com/tempura-shrimp/co/sync"
)

type AsyncChannel[R any] struct {
	*asyncSequence[R]

	sourceCh chan R
}

func FromChan[R any](ch chan R) *AsyncChannel[R] {
	a := &AsyncChannel[R]{
		sourceCh: ch,
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (a *AsyncChannel[R]) Complete() *AsyncChannel[R] {
	co_sync.SafeClose(a.sourceCh)
	return a
}

func (a *AsyncChannel[R]) iterator() Iterator[R] {
	it := &asyncChannelIterator[R]{
		AsyncChannel: a,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncChannelIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncChannel[R]
}

func (it *asyncChannelIterator[R]) next() *Optional[R] {
	val, ok := <-it.sourceCh
	if !ok {
		return NewOptionalEmpty[R]()
	}

	return OptionalOf(val)
}
