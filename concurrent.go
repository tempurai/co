package co

import "sync"

type Result[R any] struct {
	Handler func() (R, error)
	Data    R
	Error   error
}

func NewResult[R any]() *Result[R] {
	return &Result[R]{}
}

type Concurrent[R any] struct {
	Results []*Result[R]

	mux sync.Mutex
}

func NewConcurrent[R any]() *Concurrent[R] {
	return &Concurrent[R]{
		Results: make([]*Result[R], 0),
	}
}

func (co *Concurrent[R]) Len() int {
	return len(co.Results)
}

func (co *Concurrent[R]) Insert(idx int, resp *Result[R]) {
	co.mux.Lock()
	defer co.mux.Unlock()

	co.Results[idx] = resp
}

func (co *Concurrent[R]) InsertVal(idx int, val R) {
	co.mux.Lock()
	defer co.mux.Unlock()

	co.Results[idx].Data = val
}

func (co *Concurrent[R]) InsertError(idx int, err error) {
	co.mux.Lock()
	defer co.mux.Unlock()

	co.Results[idx].Error = err
}

func (co *Concurrent[R]) Append(resp ...*Result[R]) {
	co.mux.Lock()
	defer co.mux.Unlock()

	co.Results = append(co.Results, resp...)
}

func (co *Concurrent[R]) Resize(newSize int) {
	co.mux.Lock()
	defer co.mux.Unlock()

	currentSize := len(co.Results)
	if newSize < currentSize {
		return
	}

	co.Results = append(co.Results, make([]*Result[R], newSize-currentSize)...)
}

func (co *Concurrent[R]) First() *Result[R] {
	if len(co.Results) == 0 {
		return NewResult[R]()
	}
	return co.Results[0]
}

func (co *Concurrent[R]) GetAll() []*Result[R] {
	return co.Results
}

func (co *Concurrent[R]) GetAllData() []R {
	data := make([]R, len(co.Results))
	for i := range co.Results {
		data = append(data, co.Results[i].Data)
	}
	return data
}