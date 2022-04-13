package co

type AsyncSequenceable[R any] interface {
	iterator() Iterator[R]
	Iter() <-chan *data[R]
}

func toAsyncIterators[R any](cos ...AsyncSequenceable[R]) []Iterator[R] {
	its := make([]Iterator[R], len(cos))
	for i := range cos {
		its = append(its, cos[i].iterator())
	}
	return its
}
