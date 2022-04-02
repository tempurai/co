package co

type IteratorAction[T any] interface {
	next() (*Optional[T], error)
	Emitter() <-chan *data[T]
}

type IteratorAnyAction interface {
	nextAny() (*Optional[any], error)
}

type IteratorOperator interface {
}

type Iterator[T any] interface {
	IteratorAction[T]
	IteratorOperator
}

type IteratorAny interface {
	IteratorAnyAction
	IteratorOperator
}

func castToIteratorAny(vals ...any) []IteratorAny {
	casted := make([]IteratorAny, len(vals))
	for i := range vals {
		casted[i] = vals[i].(IteratorAny)
	}
	return casted
}
