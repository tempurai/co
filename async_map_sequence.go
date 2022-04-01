package co

type AsyncMapSequence[R, T any] struct {
	*asyncSequence[T]

	previousIterator Iterator[R]
	predictorFn      func(R) T
}

func NewAsyncMapSequence[R, T any](p AsyncSequenceable[R], fn func(R) T) *AsyncMapSequence[R, T] {
	a := &AsyncMapSequence[R, T]{
		previousIterator: p.Iterator(),
		predictorFn:      fn,
	}
	a.asyncSequence = NewAsyncSequence[T](a)
	return a
}

func (c *AsyncMapSequence[R, T]) SetPredicator(fn func(R) T) *AsyncMapSequence[R, T] {
	c.predictorFn = fn
	return c
}

func (a *AsyncMapSequence[R, T]) Iterator() Iterator[T] {
	it := &asyncMapSequenceIterator[R, T]{
		AsyncMapSequence: a,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[T](it)
	return it
}

type asyncMapSequenceIterator[R, T any] struct {
	*asyncSequenceIterator[T]

	*AsyncMapSequence[R, T]
	preProcessed bool
}

func (it *asyncMapSequenceIterator[R, T]) preflight() bool {
	defer func() { it.preProcessed = true }()
	return it.previousIterator.preflight()
}

func (it *asyncMapSequenceIterator[R, T]) consume() (T, error) {
	if !it.preProcessed {
		it.preflight()
	}
	defer func() { it.preProcessed = false }()

	val, err := it.previousIterator.consume()
	if err != nil {
		return *new(T), err
	}

	return it.predictorFn(val), err
}

func (it *asyncMapSequenceIterator[R, T]) next() (T, error) {
	it.preflight()
	return it.consume()
}
