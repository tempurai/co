package co

type AsyncChannel[R any] struct {
	*asyncSequence[R]

	sourceCh chan R
}

func OfChannel[R any](ch chan R) *AsyncChannel[R] {
	a := &AsyncChannel[R]{
		sourceCh: ch,
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (a *AsyncChannel[R]) Done() *AsyncChannel[R] {
	SafeClose(a.sourceCh)
	return a
}

func (a *AsyncChannel[R]) Iterator() Iterator[R] {
	it := &asyncChannelIterator[R]{
		AsyncChannel: a,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncChannelIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncChannel[R]
}

func (it *asyncChannelIterator[R]) next() (*Optional[R], error) {
	val, ok := <-it.sourceCh
	if !ok {
		return NewOptionalEmpty[R](), nil
	}

	return OptionalOf(val), nil
}
