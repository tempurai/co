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

func (a *Await[R]) Add(fns ...func() (R, error)) *Await[R] {
	respSlice := make([]*Result[R], len(fns))
	for i, fn := range fns {
		respSlice[i] = NewResult[R]()
		respSlice[i].SetExe(fn)
	}

	a.co.Append(respSlice...)
	return a
}

func (a *Await[R]) All() []*Result[R] {
	wg := sync.WaitGroup{}
	wg.Add(len(a.co.Results))

	for i := range a.co.Results {
		go func(i int) {
			a.co.Results[i].Data, a.co.Results[i].Error = a.co.Results[i].Exe()
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

			val, err := a.co.Results[i].Exe()
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

func AwaitAll[R any](fns ...func() (R, error)) []*Result[R] {
	return NewAwait[R]().Add(fns...).All()
}

func AwaitRace[R any](fns ...func() (R, error)) R {
	return NewAwait[R]().Add(fns...).Race()
}

func AwaitAny[R any](fns ...func() (R, error)) R {
	return NewAwait[R]().Add(fns...).Any()
}
