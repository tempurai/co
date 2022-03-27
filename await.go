package co

import (
	"sync"
)

type Await[R any] struct {
	co *Concurrent[R]

	raceIncludeError bool
}

func NewAwait[R any]() *Await[R] {
	return &Await[R]{
		co: NewConcurrent[R](),
	}
}

func (a *Await[R]) Add(handlers ...func() (R, error)) *Await[R] {
	respSlice := make([]*Result[R], len(handlers))
	for i, fn := range handlers {
		resp := NewResult[R]()
		resp.Handler = fn
		respSlice[i] = resp
	}

	a.co.Append(respSlice...)
	return a
}

func (a *Await[R]) All() []*Result[R] {
	wg := sync.WaitGroup{}
	wg.Add(len(a.co.Results))

	for i := range a.co.Results {
		go func(i int) {
			a.co.Results[i].Data, a.co.Results[i].Error = a.co.Results[i].Handler()
			wg.Done()
		}(i)
	}

	wg.Wait()
	return a.co.GetAll()
}

func (a *Await[R]) Race() R {
	resp := make(chan R)
	aBool := &AtomicBool{}

	for i := range a.co.Results {
		go func(i int) {
			if aBool.Get() {
				return
			}

			val, err := a.co.Results[i].Handler()
			if err != nil && !a.raceIncludeError {
				return
			}

			SafeSend(resp, val)
			aBool.Set(true)
		}(i)
	}

	val := <-resp
	close(resp)

	return val
}

func (a *Await[R]) Any() R {
	a.raceIncludeError = true
	return a.Race()
}

func AwaitAll[R any](handlers ...func() (R, error)) []*Result[R] {
	return NewAwait[R]().Add(handlers...).All()
}

func AwaitRace[R any](handlers ...func() (R, error)) R {
	return NewAwait[R]().Add(handlers...).Race()
}

func AwaitAny[R any](handlers ...func() (R, error)) R {
	return NewAwait[R]().Add(handlers...).Any()
}
