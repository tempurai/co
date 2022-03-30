package co

type IteratorAction[T any] interface {
	next() *executable[T]
}

type IteratorOperator interface {
	avaliable() bool // A Sequence is ready to execute next function
	hasNext() bool   // A Sequence have next executable function
}

type Iterator[T any] interface {
	IteratorAction[T]
	IteratorOperator
}

type ExecutableIterator[T any] interface {
	Iterator[T]

	exeNext() (T, error)
	exeFn() func() (T, error)
}

type AnyExecutableIterator interface {
	IteratorOperator

	exeNextAsAny() (any, error)
	exeFnAsAny() func() (any, error)
}

func castToAnyExecutableIterator(vals ...any) []AnyExecutableIterator {
	casted := make([]AnyExecutableIterator, len(vals))
	for i := range vals {
		casted[i] = vals[i].(AnyExecutableIterator)
	}

	return casted
}
