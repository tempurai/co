package co

type CoFilterSequence[R any] struct {
	previousIterator Iterator[R]

	predictorFn func(R, R) bool
}

func NewCoFilterSequence[R any](it Iterator[R]) *CoFilterSequence[R] {
	return &CoFilterSequence[R]{
		previousIterator: it,
		predictorFn:      func(_, _ R) bool { return true },
	}
}
func (c *CoFilterSequence[R]) SetPredicator(fn func(R, R) bool) *CoFilterSequence[R] {
	c.predictorFn = fn
	return c
}

func (c *CoFilterSequence[R]) Iterator() *coFilterSequenceIterator[R] {
	return &coFilterSequenceIterator[R]{
		CoFilterSequence: c,
	}
}

type coFilterSequenceIterator[R any] struct {
	*CoFilterSequence[R]

	previousData *data[R]
}

func (it *coFilterSequenceIterator[R]) hasNext() bool {
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

func (it *coFilterSequenceIterator[R]) next() (R, error) {
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

func (it *coFilterSequenceIterator[R]) nextAny() (any, error) {
	return it.next()
}
