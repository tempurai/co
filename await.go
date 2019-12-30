package co

import (
	"sync"
)

type awaitHandler func() interface{}

type awaiter struct {
	handlers  []awaitHandler
	responses []interface{}
}

func Await(handlers ...func() interface{}) []interface{} {
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
