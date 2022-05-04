package co

import syncx "go.tempura.ink/co/internal/sync"

type AsyncCompactedSequence[R comparable] struct {
	*asyncSequence[R]

	previousIterator Iterator[R]
	predictorFn      func(R) bool
}

func NewAsyncCompactedSequence[R comparable](it AsyncSequenceable[R]) *AsyncCompactedSequence[R] {
	a := &AsyncCompactedSequence[R]{
		previousIterator: it.iterator(),
		predictorFn:      func(r R) bool { return r != *new(R) },
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (c *AsyncCompactedSequence[R]) SetPredicator(fn func(R) bool) *AsyncCompactedSequence[R] {
	c.predictorFn = fn
	return c
}

func (c *AsyncCompactedSequence[R]) iterator() Iterator[R] {
	it := &asyncCompactedSequenceIterator[R]{
		AsyncCompactedSequence: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncCompactedSequenceIterator[R comparable] struct {
	*asyncSequenceIterator[R]

	*AsyncCompactedSequence[R]
}

func (it *asyncCompactedSequenceIterator[R]) next() *Optional[R] {
	for op := it.previousIterator.next(); op.valid; op = it.previousIterator.next() {
		match, err := syncx.SafeFn(func() bool {
			return it.predictorFn(op.data)
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

		if !match {
			continue
		}
		return op
	}
	return NewOptionalEmpty[R]()
}
