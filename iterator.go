package co

type Iterator[T any] interface {
	next() (*Optional[T], error)
	Iter() <-chan *data[T]
}

type iteratorAny interface {
	nextAny() (*Optional[any], error)
}

func castToIteratorAny(vals ...any) []iteratorAny {
	casted := make([]iteratorAny, len(vals))
	for i := range vals {
		casted[i] = vals[i].(iteratorAny)
	}
	return casted
}
