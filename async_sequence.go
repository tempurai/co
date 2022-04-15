package co

import (
	"sync"

	co_sync "github.com/tempura-shrimp/co/internal/sync"
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
	a._defaultIterator = a.async.iterator()
	return a._defaultIterator
}

func (a *asyncSequence[R]) Iter() <-chan R {
	return a.defaultIterator().Iter()
}

func (a *asyncSequence[R]) EIter() <-chan *data[R] {
	return a.defaultIterator().EIter()
}

func (a *asyncSequence[R]) AdjacentFilter(fn func(R, R) bool) *AsyncAdjacentFilterSequence[R] {
	return NewAsyncAdjacentFilterSequence(a.async, fn)
}

func (a *asyncSequence[R]) Merge(its ...AsyncSequenceable[R]) *AsyncMergedSequence[R] {
	its = append([]AsyncSequenceable[R]{a.async}, its...)
	return NewAsyncMergedSequence(its...)
}

type ErrorMode int

const (
	ErrorModeSkip ErrorMode = iota
	ErrorModeStop
)

func (e ErrorMode) shouldSkip() bool {
	return e == ErrorModeSkip
}

func (e ErrorMode) shouldStop() bool {
	return e == ErrorModeStop
}

type asyncSequenceIterator[T any] struct {
	delegated Iterator[T]
	mux       sync.RWMutex
	runOnce   sync.Once

	emitCh   []chan T
	emitECh  []chan *data[T]
	dataPool sync.Pool

	errorMode ErrorMode
	successFn func(T)
	failedFn  func(error)
}

func NewAsyncSequenceIterator[T any](it Iterator[T]) *asyncSequenceIterator[T] {
	a := &asyncSequenceIterator[T]{
		delegated: it,
		emitCh:    make([]chan T, 0),
		emitECh:   make([]chan *data[T], 0),
		dataPool:  sync.Pool{},
	}
	a.dataPool.New = func() any { return NewData[T]() }
	return a
}

func (it *asyncSequenceIterator[T]) nextAny() *Optional[any] {
	return it.delegated.next().AsOptional()
}

func (it *asyncSequenceIterator[T]) handleError(err error) {
	it.failedFn(err)

	d := it.dataPool.Get().(*data[T]).set(*new(T), err)
	it.emitData(d)
}

func (it *asyncSequenceIterator[T]) emitData(d *data[T]) {
	defer func() { it.dataPool.Put(d) }()

	it.mux.RLock()
	defer it.mux.RUnlock()

	for i := range it.emitECh {
		co_sync.SafeSend(it.emitECh[i], d)
	}
	if d.err != nil {
		return
	}

	for i := range it.emitCh {
		co_sync.SafeSend(it.emitCh[i], d.GetValue())
	}
}

func (it *asyncSequenceIterator[T]) startListening() {
	it.runOnce.Do(func() {
		co_sync.SafeGo(func() {
			for op := it.delegated.next(); op.valid; op = it.delegated.next() {
				d := it.dataPool.Get().(*data[T]).set(op.data, nil)
				it.emitData(d)

				if it.successFn != nil {
					co_sync.SafeGo(func() { it.successFn(op.data) })
				}
			}
			for _, ch := range it.emitCh {
				co_sync.SafeClose(ch)
			}
		})
	})
}

func (it *asyncSequenceIterator[T]) Iter() <-chan T {
	it.mux.Lock()
	defer it.mux.Unlock()

	eCh := make(chan T)
	it.emitCh = append(it.emitCh, eCh)

	it.startListening()
	return eCh
}

func (it *asyncSequenceIterator[T]) EIter() <-chan *data[T] {
	it.mux.Lock()
	defer it.mux.Unlock()

	eCh := make(chan *data[T])
	it.emitECh = append(it.emitECh, eCh)

	it.startListening()
	return eCh
}

func (it *asyncSequenceIterator[T]) ForEach(success func(T), failed func(error)) {
	it.successFn = success
	it.failedFn = failed
	it.startListening()
}
