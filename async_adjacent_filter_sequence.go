package co

type AsyncAdjacentFilterSequence[R any] struct {
	*asyncSequence[R]

	previousIterator Iterator[R]
	predictorFn      func(R, R) bool
}

func NewAsyncAdjacentFilterSequence[R any](it AsyncSequenceable[R], fn func(R, R) bool) *AsyncAdjacentFilterSequence[R] {
	a := &AsyncAdjacentFilterSequence[R]{
		previousIterator: it.Iterator(),
		predictorFn:      fn,
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (c *AsyncAdjacentFilterSequence[R]) SetPredicator(fn func(R, R) bool) *AsyncAdjacentFilterSequence[R] {
	c.predictorFn = fn
	return c
}

func (c *AsyncAdjacentFilterSequence[R]) Iterator() Iterator[R] {
	it := &asyncAdjacentFilterSequenceIterator[R]{
		AsyncAdjacentFilterSequence: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncAdjacentFilterSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncAdjacentFilterSequence[R]

	previousData *R
}

func (it *asyncAdjacentFilterSequenceIterator[R]) preflight() {
	if it.previousData != nil {
		return
	}

	for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
		if err != nil {
			continue
		}
		it.previousData = &op.data
		break
	}
}

func (it *asyncAdjacentFilterSequenceIterator[R]) next() (*Optional[R], error) {
	it.preflight()

	if it.previousData == nil {
		return NewOptionalEmpty[R](), nil
	}

	previousData := *it.previousData
	for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
		if err != nil {
			continue
		}
		it.previousData = &op.data
		if it.predictorFn(previousData, op.data) {
			return OptionalOf(previousData), nil
		}
	}

	it.previousData = nil
	return OptionalOf(previousData), nil
}
