package co

type AsyncFilterSequence[R any] struct {
	*asyncSequence[R]

	previousIterator Iterator[R]
	predictorFn      func(R, R) bool
}

func NewAsyncFilterSequence[R any](it AsyncSequenceable[R]) *AsyncFilterSequence[R] {
	a := &AsyncFilterSequence[R]{
		previousIterator: it.Iterator(),
		predictorFn:      func(_, _ R) bool { return true },
	}
	a.asyncSequence = NewAsyncSequence(a.Iterator())
	return a
}

func (c *AsyncFilterSequence[R]) SetPredicator(fn func(R, R) bool) *AsyncFilterSequence[R] {
	c.predictorFn = fn
	return c
}

func (c *AsyncFilterSequence[R]) Iterator() Iterator[R] {
	it := &asyncFilterSequenceIterator[R]{
		AsyncFilterSequence: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncFilterSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncFilterSequence[R]

	preProcessed bool
	previousData *data[R]
}

func (it *asyncFilterSequenceIterator[R]) preflight() bool {
	defer func() { it.preProcessed = true }()

	if it.previousData == nil && !it.previousIterator.preflight() {
		return false
	}
	if it.previousData != nil && !it.previousIterator.preflight() {
		return true
	}
	if it.previousData == nil && it.previousIterator.preflight() {
		val, err := it.previousIterator.consume()
		it.previousData = NewDataWith(val, err)
		return true
	}
	return false
}

func (it *asyncFilterSequenceIterator[R]) consume() (R, error) {
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
		if it.predictorFn(it.previousData.value, val) {
			return rData.value, rData.err
		}
	}
	it.previousData = nil
	return rData.value, rData.err
}

func (it *asyncFilterSequenceIterator[R]) next() (R, error) {
	it.preflight()
	return it.consume()
}
