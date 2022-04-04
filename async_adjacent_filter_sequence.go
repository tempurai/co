package co

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

	for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
		if err != nil {
			continue
		}
		it.previousData = *op
		break
	}
}

func (it *asyncAdjacentFilterSequenceIterator[R]) next() (*Optional[R], error) {
	it.preflight()

	if !it.previousData.valid {
		return NewOptionalEmpty[R](), nil
	}

	previousValue := it.previousData.data
	for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
		if err != nil {
			continue
		}
		it.previousData = *op
		if it.predictorFn(previousValue, op.data) {
			return OptionalOf(previousValue), nil
		}
	}

	it.previousData = *NewOptionalEmpty[R]()
	return OptionalOf(previousValue), nil
}
