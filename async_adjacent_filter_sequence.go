package co

import (
	co_sync "github.com/tempura-shrimp/co/sync"
)

type AsyncAdjacentFilterSequence[R any] struct {
	*asyncSequence[R]

	previousIterator Iterator[R]
	predictorFn      func(R, R) bool
}

func NewAsyncAdjacentFilterSequence[R any](it AsyncSequenceable[R], fn func(R, R) bool) *AsyncAdjacentFilterSequence[R] {
	a := &AsyncAdjacentFilterSequence[R]{
		previousIterator: it.iterator(),
		predictorFn:      fn,
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (c *AsyncAdjacentFilterSequence[R]) SetPredicator(fn func(R, R) bool) *AsyncAdjacentFilterSequence[R] {
	c.predictorFn = fn
	return c
}

func (c *AsyncAdjacentFilterSequence[R]) iterator() Iterator[R] {
	it := &asyncAdjacentFilterSequenceIterator[R]{
		AsyncAdjacentFilterSequence: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncAdjacentFilterSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncAdjacentFilterSequence[R]
	previousData Optional[R]
}

func (it *asyncAdjacentFilterSequenceIterator[R]) preflight() {
	if it.previousData.valid {
		return
	}

	for op := it.previousIterator.next(); op.valid; op = it.previousIterator.next() {
		it.previousData = *op
		break
	}
}

func (it *asyncAdjacentFilterSequenceIterator[R]) next() *Optional[R] {
	it.preflight()

	if !it.previousData.valid {
		return NewOptionalEmpty[R]()
	}

	previousValue := it.previousData.data
	for op := it.previousIterator.next(); op.valid; op = it.previousIterator.next() {
		it.previousData = *op

		match, err := co_sync.SafeFn(func() bool {
			return it.predictorFn(previousValue, op.data)
		})
		if err != nil {
			it.handleError(err)
			if it.errorMode.shouldSkip() {
				continue
			}
			if it.errorMode.shouldStop() {
				return NewOptionalEmpty[R]()
			}
		}
		if match {
			return OptionalOf(previousValue)
		}
	}

	it.previousData = *NewOptionalEmpty[R]()
	return OptionalOf(previousValue)
}
