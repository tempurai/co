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

	preProcessed bool
	previousData *data[R]
}

func (it *asyncAdjacentFilterSequenceIterator[R]) preflight() bool {
	defer func() { it.preProcessed = true }()

	if it.previousData != nil {
		return true
	}
	if it.previousData == nil && !it.previousIterator.preflight() {
		return false
	}
	if it.previousData == nil && it.previousIterator.preflight() {
		val, err := it.previousIterator.consume()
		it.previousData = NewDataWith(val, err)
		return true
	}
	return false
}

func (it *asyncAdjacentFilterSequenceIterator[R]) consume() (R, error) {
	if !it.preProcessed {
		it.preflight()
	}
	defer func() { it.preProcessed = false }()

	rData := it.previousData
	for it.previousIterator.preflight() {
		val, err := it.previousIterator.consume()
		it.previousData = NewDataWith(val, err)
		if err != nil {
			return val, err
		}
		if it.predictorFn(rData.value, it.previousData.value) {
			return rData.value, rData.err
		}
	}
	it.previousData = nil
	return rData.value, rData.err
}
