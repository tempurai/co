package co

type AsyncPartitionSequence[R any] struct {
	*asyncSequence[[]R]

	previousIterators []Iterator[R]
	size              int
}

func NewAsyncPartitionSequence[R any](size int, its ...AsyncSequenceable[R]) *AsyncPartitionSequence[R] {
	a := &AsyncPartitionSequence[R]{
		previousIterators: toAsyncIterators(its...),
		size:              size,
	}
	a.asyncSequence = NewAsyncSequence[[]R](a)
	return a
}

func (a *AsyncPartitionSequence[R]) SetSize(size int) *AsyncPartitionSequence[R] {
	a.size = size
	return a
}

func (c *AsyncPartitionSequence[R]) iterator() Iterator[[]R] {
	it := &asyncPartitionSequenceIterator[R, []R]{
		AsyncPartitionSequence: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[[]R](it)
	return it
}

type asyncPartitionSequenceIterator[T any, R []T] struct {
	*asyncSequenceIterator[R]

	*AsyncPartitionSequence[T]
	currentIndex int
}

func (it *asyncPartitionSequenceIterator[T, R]) nextIndex() int {
	defer func() {
		if it.currentIndex+1 >= len(it.previousIterators) {
			it.currentIndex = 0
		} else {
			it.currentIndex++
		}
	}()
	return it.currentIndex
}

func (it *asyncPartitionSequenceIterator[T, R]) getNextPrevious() *Optional[T] {
	return it.previousIterators[it.nextIndex()].next()
}

func (it *asyncPartitionSequenceIterator[T, R]) next() *Optional[R] {
	results := make(R, 0)

	for op := it.getNextPrevious(); op.valid; op = it.getNextPrevious() {
		results = append(results, op.data)
		if len(results) == it.size {
			break
		}
	}

	if len(results) == 0 {
		return NewOptionalEmpty[R]()
	}

	return OptionalOf(results)
}
