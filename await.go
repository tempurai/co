package co

import (
	"sync"
)

type awaitHandler func() interface{}

type awaiter struct {
	handlers  []awaitHandler
	responses []interface{}
}

func AwaitAll(handlers ...func() interface{}) []interface{} {
	w := &awaiter{
		handlers:  make([]awaitHandler, len(handlers)),
		responses: make([]interface{}, len(handlers)),
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

func AwaitRace(handlers ...func() (interface{}, error)) interface{} {
	resp := make(chan interface{})

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

func AwaitAny(handlers ...func() interface{}) interface{} {
	wrappedHandlers := make([]func() (interface{}, error), 0)

	for i := range handlers {
		func(i int) {
			wrappedHandlers[i] = func() (interface{}, error) {
				return handlers[i](), nil
			}
		}(i)
	}

	return AwaitRace(wrappedHandlers...)
}
