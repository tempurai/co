package co

type AsyncCompactedSequence[R comparable] struct {
	previousIterator Iterator[R]

	predictorFn func(R) bool
}

func NewAsyncCompactedSequence[R comparable](it Iterator[R]) *AsyncCompactedSequence[R] {
	return &AsyncCompactedSequence[R]{
		previousIterator: it,
		predictorFn:      func(r R) bool { return r != *new(R) },
	}
}

func (c *AsyncCompactedSequence[R]) Iterator() *asyncCompactedSequenceIterator[R] {
	return &asyncCompactedSequenceIterator[R]{
		AsyncCompactedSequence: c,
	}
}

type asyncCompactedSequenceIterator[R comparable] struct {
	*AsyncCompactedSequence[R]

	preProcessed bool
	previousData *data[R]
}

func (it *asyncCompactedSequenceIterator[R]) hasNext() bool {
	defer func() { it.preProcessed = true }()

	if it.previousData == nil && !it.previousIterator.hasNext() {
		return false
	}
	if it.previousData != nil && !it.previousIterator.hasNext() {
		return true
	}
	if it.previousData == nil && it.previousIterator.hasNext() {
		for it.previousIterator.hasNext() {
			val, err := it.previousIterator.consume()
			if err != nil {
				it.previousData = NewDataWith(val, err)
				return true
			}

			if !it.predictorFn(val) {
				continue
			}

			it.previousData = NewDataWith(val, err)
			return true
		}
	}
	return false
}

func (it *asyncCompactedSequenceIterator[R]) consume() (R, error) {
	if !it.preProcessed {
		it.hasNext()
	}
	defer func() { it.preProcessed = false }()

	rData := it.previousData
	it.previousData = nil
	return rData.value, rData.err
}

func (it *asyncCompactedSequenceIterator[R]) consumeAny() (any, error) {
	return it.consume()
}

func (it *asyncCompactedSequenceIterator[R]) next() (R, error) {
	it.hasNext()
	return it.consume()
}
