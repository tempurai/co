package co

type Concurrently[R any] interface {
	Iterator() ExecutableIterator[R]
}

func toConcurrentIterators[R any](cos ...Concurrently[R]) []ExecutableIterator[R] {
	its := make([]ExecutableIterator[R], len(cos))
	for i := range cos {
		its = append(its, cos[i].Iterator())
	}
	return its
}
