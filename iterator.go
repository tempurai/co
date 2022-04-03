package co

type iteratorAction[T any] interface {
	next() (*Optional[T], error)
	Emitter() <-chan *data[T]
}

type iteratorAnyAction interface {
	nextAny() (*Optional[any], error)
}

type iteratorOperator interface {
}

type Iterator[T any] interface {
	iteratorAction[T]
	iteratorOperator
}

type iteratorAny interface {
	iteratorAnyAction
	iteratorOperator
}

func castToIteratorAny(vals ...any) []iteratorAny {
	casted := make([]iteratorAny, len(vals))
	for i := range vals {
		casted[i] = vals[i].(iteratorAny)
	}
	return casted
}
