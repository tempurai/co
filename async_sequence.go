package co

import (
	"sync"
)

type asyncSequenceIterator[R, T any] struct {
	delegated Iterator[T]

	emitCh        []chan *data[T]
	isEmitRunning bool
	emitMux       sync.Mutex
}

func NewAsyncSequenceIterator[R, T any](it Iterator[T]) *asyncSequenceIterator[R, T] {
	return &asyncSequenceIterator[R, T]{delegated: it, emitCh: make([]chan *data[T], 0)}
}

func (it *asyncSequenceIterator[R, T]) consumeAny() (any, error) {
	return it.delegated.consume()
}

func (it *asyncSequenceIterator[R, T]) next() (T, error) {
	it.delegated.preflight()
	return it.delegated.consume()
}

func (it *asyncSequenceIterator[R, T]) emitData(d *data[T]) {
	it.emitMux.Lock()
	defer it.emitMux.Unlock()

	for _, ch := range it.emitCh {
		SafeSend(ch, d)
	}
}

func (it *asyncSequenceIterator[R, T]) emitIterator() {
	if it.isEmitRunning {
		return
	}
	it.isEmitRunning = true

	SafeGo(func() {
		for it.delegated.preflight() {
			val, err := it.delegated.consume()
			it.emitData(NewDataWith(val, err))
		}
		for _, ch := range it.emitCh {
			SafeClose(ch)
		}
	})
}

func (it *asyncSequenceIterator[R, T]) EmitIterator() <-chan *data[T] {
	it.emitMux.Lock()
	defer it.emitMux.Unlock()

	eCh := make(chan *data[T])
	it.emitCh = append(it.emitCh, eCh)

	it.emitIterator()
	return eCh
}
