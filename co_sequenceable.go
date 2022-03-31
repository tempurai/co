package co

type CoSequenceable[R any] interface {
	Iterator() Iterator[R]
}

func toConcurrentIterators[R any](cos ...CoSequenceable[R]) []Iterator[R] {
	its := make([]Iterator[R], len(cos))
	for i := range cos {
		its = append(its, cos[i].Iterator())
	}
	return its
}
