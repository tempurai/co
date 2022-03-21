package co

import (
	"sync"
)

type awaitHandler[K any] func() K

type awaiter[T any] struct {
	handlers  []awaitHandler[T]
	responses []T
}

func AwaitAll[K any](handlers ...func() K) []K {
	w := &awaiter[K]{
		handlers:  make([]awaitHandler[K], len(handlers)),
		responses: make([]K, len(handlers)),
	}

	wg := sync.WaitGroup{}
	wg.Add(len(handlers))

	for i := range handlers {
		w.handlers[i] = handlers[i]

		go func(i int) {
			vals := w.handlers[i]()
			w.responses[i] = vals

			wg.Done()
		}(i)
	}

	wg.Wait()
	return w.responses
}

func AwaitRace[K any](handlers ...func() (K, error)) K {
	resp := make(chan K)

	for i := range handlers {
		go func(i int) {
			val, err := handlers[i]()
			if err != nil {
				return
			}
			resp <- val
		}(i)
	}

	return <-resp
}

func AwaitAny[K any](handlers ...func() K) K {
	wrappedHandlers := make([]func() (K, error), 0)

	for i := range handlers {
		func(i int) {
			wrappedHandlers[i] = func() (K, error) {
				return handlers[i](), nil
			}
		}(i)
	}

	return AwaitRace(wrappedHandlers...)
}
