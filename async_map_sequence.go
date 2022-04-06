package co

type AsyncMapSequence[R, T any] struct {
	*asyncSequence[T]

	previousIterator Iterator[R]
	predictorFn      func(R) T
}

func NewAsyncMapSequence[R, T any](p AsyncSequenceable[R], fn func(R) T) *AsyncMapSequence[R, T] {
	a := &AsyncMapSequence[R, T]{
		previousIterator: p.iterator(),
		predictorFn:      fn,
	}
	a.asyncSequence = NewAsyncSequence[T](a)
	return a
}

func (c *AsyncMapSequence[R, T]) SetPredicator(fn func(R) T) *AsyncMapSequence[R, T] {
	c.predictorFn = fn
	return c
}

func (a *AsyncMapSequence[R, T]) iterator() Iterator[T] {
	it := &asyncMapSequenceIterator[R, T]{
		AsyncMapSequence: a,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[T](it)
	return it
}

type asyncMapSequenceIterator[R, T any] struct {
	*asyncSequenceIterator[T]

	*AsyncMapSequence[R, T]
}

func (it *asyncMapSequenceIterator[R, T]) next() (*Optional[T], error) {
	for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
		if err != nil {
			return nil, err
		}
		return OptionalOf(it.predictorFn(op.data)), nil
	}
	return NewOptionalEmpty[T](), nil
}
