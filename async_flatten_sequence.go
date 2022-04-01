package co

type AsyncFlattenSequence[R any, T []R] struct {
	*asyncSequence[R]

	previousIterator Iterator[T]
}

func NewAsyncFlattenSequence[R any, T []R](it AsyncSequenceable[T]) *AsyncFlattenSequence[R, T] {
	a := &AsyncFlattenSequence[R, T]{
		previousIterator: it.Iterator(),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (c *AsyncFlattenSequence[R, T]) Iterator() Iterator[R] {
	it := &asyncFlattenSequenceIterator[R, T]{
		AsyncFlattenSequence: c,
		bufferedData:         make(T, 0),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncFlattenSequenceIterator[R any, T []R] struct {
	*asyncSequenceIterator[R]

	*AsyncFlattenSequence[R, T]
	preProcessed bool
	bufferedData T // Should to optimze with linked list
}

func (it *asyncFlattenSequenceIterator[R, T]) preflight() bool {
	defer func() { it.preProcessed = true }()
	if len(it.bufferedData) > 0 {
		return true
	}

	for it.previousIterator.preflight() {
		val, _ := it.previousIterator.consume()
		if len(val) > 0 {
			it.bufferedData = append(it.bufferedData, val...)
			return true
		}
	}

	return false
}

func (it *asyncFlattenSequenceIterator[R, T]) consume() (R, error) {
	if !it.preProcessed {
		it.preflight()
	}
	defer func() { it.preProcessed = false }()

	result, nextBuffer := it.bufferedData[0], it.bufferedData[1:]
	it.bufferedData = nextBuffer

	return result, nil
}

func (it *asyncFlattenSequenceIterator[R, T]) next() (R, error) {
	it.preflight()
	return it.consume()
}
