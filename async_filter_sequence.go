package co

type AsyncFilterSequence[R any] struct {
	previousIterator Iterator[R]

	predictorFn func(R, R) bool
}

func NewAsyncFilterSequence[R any](it Iterator[R]) *AsyncFilterSequence[R] {
	return &AsyncFilterSequence[R]{
		previousIterator: it,
		predictorFn:      func(_, _ R) bool { return true },
	}
}
func (c *AsyncFilterSequence[R]) SetPredicator(fn func(R, R) bool) *AsyncFilterSequence[R] {
	c.predictorFn = fn
	return c
}

func (c *AsyncFilterSequence[R]) Iterator() *asyncFilterSequenceIterator[R] {
	return &asyncFilterSequenceIterator[R]{
		AsyncFilterSequence: c,
	}
}

type asyncFilterSequenceIterator[R any] struct {
	*AsyncFilterSequence[R]

	previousData *data[R]
}

func (it *asyncFilterSequenceIterator[R]) hasNext() bool {
	if it.previousData == nil && !it.previousIterator.hasNext() {
		return false
	}
	if it.previousData != nil && !it.previousIterator.hasNext() {
		return true
	}
	if it.previousData == nil && it.previousIterator.hasNext() {
		val, err := it.previousIterator.next()
		it.previousData = NewDataWith(val, err)
		return true
	}
	return false
}

func (it *asyncFilterSequenceIterator[R]) next() (R, error) {
	for it.previousIterator.hasNext() {
		for it.previousIterator.hasNext() {
			val, err := it.previousIterator.next()
			if err != nil {
				it.previousData = NewDataWith(val, err)
				return val, err
			}
			if it.predictorFn(it.previousData.value, val) {
				rData := it.previousData
				it.previousData = NewDataWith(val, err)
				return rData.value, rData.err
			}
		}
	}
	rData := it.previousData
	it.previousData = nil
	return rData.value, rData.err
}

func (it *asyncFilterSequenceIterator[R]) nextAny() (any, error) {
	return it.next()
}
