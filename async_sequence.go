package co

type asyncSequenceIterator[R, T any] struct {
	delegated Iterator[T]
}

func NewAsyncSequenceIterator[R, T any](it Iterator[T]) *asyncSequenceIterator[R, T] {
	return &asyncSequenceIterator[R, T]{delegated: it}
}

func (it *asyncSequenceIterator[R, T]) consumeAny() (any, error) {
	return it.delegated.consume()
}

func (it *asyncSequenceIterator[R, T]) next() (T, error) {
	it.delegated.preflight()
	return it.delegated.consume()
}
