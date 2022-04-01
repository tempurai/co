package co

import (
	"sync"
)

type asyncSequence[R, T any] struct {
	delegated AsyncSequenceable[R]

	_defaultIterator Iterator[T]
}

func NewAsyncSequence[R, T any](it AsyncSequenceable[R]) *asyncSequence[R, T] {
	return &asyncSequence[R, T]{delegated: it}
}

// func (a *asyncSequence[R, T]) defaultIterator() Iterator[T] {
// 	if a._defaultIterator != nil {
// 		return a._defaultIterator
// 	}
// 	a._defaultIterator = a.delegated.Iterator().Emit()
// 	return a._defaultIterator
// }

// func (a *asyncSequence[R, T]) Emitter() <-chan *data[T] {
// 	return a.delegated.Iterator().Emit()
// }

type asyncSequenceIterator[T any] struct {
	delegated Iterator[T]

	emitCh        []chan *data[T]
	isEmitRunning bool
	emitMux       sync.Mutex
}

func NewAsyncSequenceIterator[T any](it Iterator[T]) *asyncSequenceIterator[T] {
	return &asyncSequenceIterator[T]{delegated: it, emitCh: make([]chan *data[T], 0)}
}

func (it *asyncSequenceIterator[T]) consumeAny() (any, error) {
	return it.delegated.consume()
}

func (it *asyncSequenceIterator[T]) next() (T, error) {
	it.delegated.preflight()
	return it.delegated.consume()
}

func (it *asyncSequenceIterator[T]) emitData(d *data[T]) {
	it.emitMux.Lock()
	defer it.emitMux.Unlock()

	for _, ch := range it.emitCh {
		SafeSend(ch, d)
	}
}

func (it *asyncSequenceIterator[T]) runEmit() {
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

func (it *asyncSequenceIterator[T]) Emitter() <-chan *data[T] {
	it.emitMux.Lock()
	defer it.emitMux.Unlock()

	eCh := make(chan *data[T])
	it.emitCh = append(it.emitCh, eCh)

	it.runEmit()
	return eCh
}
