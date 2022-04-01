package co

type IteratorAction[T any] interface {
	consume() (T, error)
	next() (T, error)
}

type IteratorAnyAction interface {
	consumeAny() (any, error)
}

type IteratorOperator interface {
	hasNext() bool // A Sequence have consume executable function
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
