package co

type IteratorAction[T any] interface {
	next() (T, error)
}

type IteratorAnyAction interface {
	nextAny() (any, error)
}

type IteratorOperator interface {
	hasNext() bool // A Sequence have next executable function
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
