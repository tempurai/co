package co

type AsyncSequenceable[R any] interface {
	iterator() Iterator[R]
}

func toAsyncIterators[R any](cos ...AsyncSequenceable[R]) []Iterator[R] {
	its := make([]Iterator[R], len(cos))
	for i := range cos {
		its[i] = cos[i].iterator()
	}
	return its
}
