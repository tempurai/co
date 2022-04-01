package co

type asyncSequenceIterator[R any] struct {
	delegated Iterator[R]
}

func NewAsyncSequenceIterator[R any](it Iterator[R]) *asyncSequenceIterator[R] {
	return &asyncSequenceIterator[R]{delegated: it}
}

func (it *asyncSequenceIterator[R]) consumeAny() (any, error) {
	return it.delegated.consume()
}

func (it *asyncSequenceIterator[R]) next() (R, error) {
	it.delegated.preflight()
	return it.delegated.consume()
}
