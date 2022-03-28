package co

type Concurrent[R any] struct {
	executorList[R]
}

func NewConcurrent[R any]() *Concurrent[R] {
	return &Concurrent[R]{
		executorList: *NewExecutorList[R](),
	}
}

func NewDataConcurrent[R any](data ...R) *Concurrent[R] {
	co := NewConcurrent[R]()

	executors := make([]*executor[R], len(data))
	for i := range data {
		executors[i].Data = data[i]
		executors[i].executed = true
	}

	co.executorAppend(executors...)
	return co
}

func NewJobConcurrent[R any](fns ...func() (R, error)) *Concurrent[R] {
	co := NewConcurrent[R]()
	co.Add(fns...)
	return co
}

func (co *Concurrent[R]) Add(fns ...func() (R, error)) *Concurrent[R] {
	respSlice := make([]*executor[R], len(fns))
	for i, fn := range fns {
		respSlice[i] = NewExecutor[R]()
		respSlice[i].setExeFn(fn)
	}

	co.executorAppend(respSlice...)
	return co
}

func (co *Concurrent[R]) Append(co2 *Concurrent[R]) *Concurrent[R] {
	co.executorAppend(co2.executors...)
	return co
}

type ConcurrentExecutor interface {
	latestFnToAny() (any, error)
}

func (co *Concurrent[R]) latestFnToAny() (any, error) {
	return co.executors[len(co.executors)-1].exe()
}
