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
	return a.defaultIterator().Emitter()
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
	mux       sync.RWMutex
	runOnce   sync.Once

	emitCh []chan *data[T]

	successFn func(T)
	failedFn  func(error)
}

func NewAsyncSequenceIterator[T any](it Iterator[T]) *asyncSequenceIterator[T] {
	return &asyncSequenceIterator[T]{
		delegated: it,
		emitCh:    make([]chan *data[T], 0),

		successFn: func(t T) {},
		failedFn:  func(err error) {},
	}
}

func (it *asyncSequenceIterator[T]) nextAny() (*Optional[any], error) {
	oVal, err := it.delegated.next()
	return oVal.AsOptional(), err
}

func (it *asyncSequenceIterator[T]) emitData(d *data[T]) {
	it.mux.RLock()
	defer it.mux.RUnlock()

	for i := range it.emitCh {
		SafeSend(it.emitCh[i], d)
	}
}

func (it *asyncSequenceIterator[T]) run() {
	it.runOnce.Do(func() {
		SafeGo(func() {
			for op, err := it.delegated.next(); op.valid; op, err = it.delegated.next() {
				// Channel
				it.emitData(NewDataWith(op.data, err))

				// Function
				if err == nil {
					it.successFn(op.data)
				} else {
					it.failedFn(err)
				}
			}
			for _, ch := range it.emitCh {
				SafeClose(ch)
			}
		})
	})
}

func (it *asyncSequenceIterator[T]) Emitter() <-chan *data[T] {
	it.mux.Lock()
	defer it.mux.Unlock()

	eCh := make(chan *data[T])
	it.emitCh = append(it.emitCh, eCh)

	it.run()
	return eCh
}

func (it *asyncSequenceIterator[T]) ForEach(success func(T), failed func(error)) {
	it.successFn = success
	it.failedFn = failed
	it.run()
}
