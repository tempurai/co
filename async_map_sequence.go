package co

import syncx "github.com/tempurai/co/internal/syncx"

type AsyncMapSequence[R, T any] struct {
	*asyncSequence[T]

	previousIterator Iterator[R]
	predictorFn      func(R) T
}

func NewAsyncMapSequence[R, T any](p AsyncSequenceable[R], fn func(R) T) *AsyncMapSequence[R, T] {
	a := &AsyncMapSequence[R, T]{
		previousIterator: p.iterator(),
		predictorFn:      fn,
	}
	a.asyncSequence = NewAsyncSequence[T](a)
	return a
}

func (c *AsyncMapSequence[R, T]) SetPredicator(fn func(R) T) *AsyncMapSequence[R, T] {
	c.predictorFn = fn
	return c
}

func (a *AsyncMapSequence[R, T]) iterator() Iterator[T] {
	it := &asyncMapSequenceIterator[R, T]{
		AsyncMapSequence: a,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[T](it)
	return it
}

type asyncMapSequenceIterator[R, T any] struct {
	*asyncSequenceIterator[T]

	*AsyncMapSequence[R, T]
}

func (it *asyncMapSequenceIterator[R, T]) next() *Optional[T] {
	for op := it.previousIterator.next(); op.valid; op = it.previousIterator.next() {
		mapped, err := syncx.SafeFn(func() T {
			return it.predictorFn(op.data)
		})
		if err != nil {
			it.handleError(err)
			if it.errorMode.shouldSkip() {
				continue
			}
			if it.errorMode.shouldStop() {
				return NewOptionalEmpty[T]()
			}
		}
		return OptionalOf(mapped)
	}
	return NewOptionalEmpty[T]()
}
