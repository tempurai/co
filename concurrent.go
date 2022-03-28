package co

type concurrentIterator struct {
	index int
}

type Concurrent[R any] struct {
	executorList[R]

	concurrentIterator concurrentIterator
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
	len() int
	exeFnAt(int) (any, error)

	prepareIterator()
	exeNextFn() (any, error)
	finished() bool
}

func (co *Concurrent[R]) len() int {
	return len(co.executors)
}

func (co *Concurrent[R]) exeFnAt(i int) (any, error) {
	return co.executors[i].exe()
}

func (co *Concurrent[R]) prepareIterator() {
	co.concurrentIterator = concurrentIterator{}
}

func (co *Concurrent[R]) exeNextFn() (any, error) {
	data, err := co.executors[co.concurrentIterator.index].exe()
	co.concurrentIterator.index++
	return data, err
}

func (co *Concurrent[R]) finished() bool {
	return co.concurrentIterator.index >= co.len()
}
