package co

import "sync"

type result[R any] struct {
	Data  R
	Error error
}

type executor[R any] struct {
	result[R]

	exeFn    func() (R, error)
	executed bool
}

func NewExecutor[R any]() *executor[R] {
	return &executor[R]{}
}

func (r *executor[R]) setExeFn(fn func() (R, error)) *executor[R] {
	r.exeFn = fn
	return r
}

func (r *executor[R]) exe() (R, error) {
	if r.isExecuted() {
		return r.Data, nil
	}
	r.Data, r.Error = r.exeFn()
	return r.Data, r.Error
}

func (r *executor[R]) forceExe() (R, error) {
	return r.exeFn()
}

func (r *executor[R]) setExecuted(b bool) {
	r.executed = b
}

func (r *executor[R]) isExecuted() bool {
	return r.executed
}

// ***************************************************************
// **************************** executors **************************
// ***************************************************************
type executorList[R any] struct {
	executors []*executor[R]

	rwmux sync.RWMutex
}

func NewExecutorList[R any]() *executorList[R] {
	return &executorList[R]{
		executors: make([]*executor[R], 0),
	}
}

func (r *executorList[R]) executorLen() int {
	return len(r.executors)
}

func (r *executorList[R]) executorInsert(idx int, resp *executor[R]) {
	r.rwmux.Lock()
	defer r.rwmux.Unlock()

	r.executors[idx] = resp
}

func (r *executorList[R]) executorAppend(resp ...*executor[R]) {
	r.rwmux.Lock()
	defer r.rwmux.Unlock()

	r.executors = append(r.executors, resp...)
}

func (r *executorList[R]) executorResize(newSize int) {
	r.rwmux.Lock()
	defer r.rwmux.Unlock()

	currentSize := len(r.executors)
	if newSize < currentSize {
		return
	}

	r.executors = append(r.executors, make([]*executor[R], newSize-currentSize)...)
}

func (r *executorList[R]) executorSwap(executor []*executor[R]) {
	r.rwmux.Lock()
	defer r.rwmux.Unlock()

	r.executors = executor
}
