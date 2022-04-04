package co

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

func (it *asyncCompactedSequenceIterator[R]) next() (*Optional[R], error) {
	for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
		if err != nil {
			return nil, err
		}
		if !it.predictorFn(op.data) {
			continue
		}
		return op, nil
	}
	return NewOptionalEmpty[R](), nil
}
