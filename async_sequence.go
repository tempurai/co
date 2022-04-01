package co

import (
	"sync"
)

type asyncSequence[R any] struct {
	async AsyncSequenceable[R]

	_defaultIterator Iterator[R]
}

func NewAsyncSequence[R any](it AsyncSequenceable[R]) *asyncSequence[R] {
	return &asyncSequence[R]{async: it}
}

func (a *asyncSequence[R]) defaultIterator() Iterator[R] {
	if a._defaultIterator != nil {
		return a._defaultIterator
	}
	a._defaultIterator = a.async.Iterator()
	return a._defaultIterator
}

func (a *asyncSequence[R]) Emitter() <-chan *data[R] {
	return a._defaultIterator.Emitter()
}

func (a *asyncSequence[R]) AdjacentFilter(fn func(R, R) bool) *AsyncAdjacentFilterSequence[R] {
	return NewAsyncAdjacentFilterSequence(a.async, fn)
}

func (a *asyncSequence[R]) Merge(its ...AsyncSequenceable[R]) *AsyncMergedSequence[R] {
	its = append([]AsyncSequenceable[R]{a.async}, its...)
	return NewAsyncMergedSequence(its...)
}

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
