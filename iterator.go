package co

type iteratorAction[T any] interface {
	next() (*Optional[T], error)
	Emitter() <-chan *data[T]
}

type iteratorAnyAction interface {
	nextAny() (*Optional[any], error)
}

type Iterator[T any] interface {
	iteratorAction[T]
}

type iteratorAny interface {
	iteratorAnyAction
}

func castToIteratorAny(vals ...any) []iteratorAny {
	casted := make([]iteratorAny, len(vals))
	for i := range vals {
		casted[i] = vals[i].(iteratorAny)
	}
	return casted
}
