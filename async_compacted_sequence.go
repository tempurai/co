package co

type AsyncCompactedSequence[R comparable] struct {
	*asyncSequence[R]

	previousIterator Iterator[R]
	predictorFn      func(R) bool
}

func NewAsyncCompactedSequence[R comparable](it AsyncSequenceable[R]) *AsyncCompactedSequence[R] {
	a := &AsyncCompactedSequence[R]{
		previousIterator: it.Iterator(),
		predictorFn:      func(r R) bool { return r != *new(R) },
	}
	a.asyncSequence = NewAsyncSequence[R](a.Iterator())
	return a
}

func (c *AsyncCompactedSequence[R]) Iterator() *asyncCompactedSequenceIterator[R] {
	it := &asyncCompactedSequenceIterator[R]{
		AsyncCompactedSequence: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncCompactedSequenceIterator[R comparable] struct {
	*asyncSequenceIterator[R]

	*AsyncCompactedSequence[R]

	preProcessed bool
	previousData *data[R]
}

func (it *asyncCompactedSequenceIterator[R]) preflight() bool {
	defer func() { it.preProcessed = true }()

	if it.previousData == nil && !it.previousIterator.preflight() {
		return false
	}
	if it.previousData != nil && !it.previousIterator.preflight() {
		return true
	}
	if it.previousData == nil && it.previousIterator.preflight() {
		for it.previousIterator.preflight() {
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
		it.preflight()
	}
	defer func() { it.preProcessed = false }()

	rData := it.previousData
	it.previousData = nil
	return rData.value, rData.err
}
