package co

type Iterator[T any] interface {
	next() *Optional[T]
	Iter() <-chan T
	EIter() <-chan *data[T]
}

type iteratorAny interface {
	nextAny() *Optional[any]
}

func castToIteratorAny(vals ...any) []iteratorAny {
	casted := make([]iteratorAny, len(vals))
	for i := range vals {
		casted[i] = vals[i].(iteratorAny)
	}
	return casted
}

type AsyncSequenceable[R any] interface {
	iterator() Iterator[R]
}

func toAsyncIterators[R any](cos ...AsyncSequenceable[R]) []Iterator[R] {
	its := make([]Iterator[R], len(cos))
	for i := range cos {
		its[i] = cos[i].iterator()
	}
	return its
}
