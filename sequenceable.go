package co

type SequenceableAction[T any] interface {
	next() *executable[T]
}

type SequenceableOperator interface {
	avaliable() bool // A Sequence is ready to execute next function
	hasNext() bool   // A Sequence have next executable function
}

type Sequenceable[T any] interface {
	SequenceableAction[T]
	SequenceableOperator
}

type ExecutableSequence[T any] interface {
	Sequenceable[T]

	exeNext() (T, error)
	exeFn() func() (T, error)
}

type GenericExecutableSequence interface {
	SequenceableOperator

	exeNextAsAny() (any, error)
	exeFnAsAny() func() (any, error)
}

func castToGenericExecutableSequence(vals ...any) []GenericExecutableSequence {
	casted := make([]GenericExecutableSequence, len(vals))
	for i := range vals {
		casted[i] = vals[i].(GenericExecutableSequence)
	}

	return casted
}

type Concurrently[R any] interface {
	Iterator() ExecutableSequence[R]
}

func toConcurrentIterators[R any](cos ...Concurrently[R]) []ExecutableSequence[R] {
	its := make([]ExecutableSequence[R], len(cos))
	for i := range cos {
		its = append(its, cos[i].Iterator())
	}
	return its
}
