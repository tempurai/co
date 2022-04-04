package co

import "sync"

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
		SafeGo(func() {
			for val := range a.sourceCh {
				for _, iCh := range a.listenerChs {
					SafeSend(iCh, val)
				}
			}
		})
	})
}

func (a *AsyncChannel[R]) Done() *AsyncChannel[R] {
	SafeClose(a.sourceCh)
	for i := range a.listenerChs {
		SafeClose(a.listenerChs[i])
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
