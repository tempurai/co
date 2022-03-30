package co

type Concurrent[R any] struct {
	executables *executablesList[R]
	data        *determinedDataList[R]

	_defaultIterator ExecutableSequence[R]
}

func NewConcurrent[R any]() *Concurrent[R] {
	return &Concurrent[R]{
		executables: NewExecutablesList[R](),
		data:        NewDeterminedDataList[R](),
	}
}

func (r *Concurrent[R]) len() int {
	return r.executables.len()
}

func (r *Concurrent[R]) exe(i int) (R, error) {
	if r.executables.getAt(i).isExecuted() {
		return r.data.getAt(i)
	}

	return r.forceExeAt(i)
}

func (r *Concurrent[R]) forceExeAt(i int) (R, error) {
	val, err := r.executables.getAt(i).exe()
	r.data.setAt(i, val, err)

	return val, err
}

func (r *Concurrent[R]) addData(dVal ...*data[R]) *Concurrent[R] {
	r.data.List.add(dVal...)

	for range dVal {
		r.executables.add(NewExecutor[R]())
	}
	return r

}

func (r *Concurrent[R]) addExeFn(fns ...func() (R, error)) *Concurrent[R] {
	for i := range fns {
		r.data.List.add(NewData[R]())

		e := NewExecutor[R]()
		e.fn = fns[i]
		r.executables.add(e)
	}
	return r
}

func (r *Concurrent[R]) add(eVal ...*executable[R]) *Concurrent[R] {
	for range eVal {
		r.data.List.add(NewData[R]())
	}

	r.executables.add(eVal...)
	return r
}

func (r *Concurrent[R]) swapValue(dVal *determinedDataList[R], eVal *executablesList[R]) *Concurrent[R] {
	if dVal != nil {
		r.data.swap(dVal.list)
	}
	if eVal != nil {
		r.executables.swap(eVal.list)
	}
	return r
}

func (r *Concurrent[R]) defaultIterator() ExecutableSequence[R] {
	if r._defaultIterator != nil {
		return r._defaultIterator
	}
	r._defaultIterator = r.Iterator()
	return r._defaultIterator
}

func (r *Concurrent[R]) Iterator() ExecutableSequence[R] {
	return &concurrentIterator[R]{
		Concurrent:            r,
		iterativeListIterator: r.executables.iterativeList.Iterator(),
	}
}

type concurrentIterator[R any] struct {
	*Concurrent[R]
	*iterativeListIterator[*executable[R]]
}

func (r *concurrentIterator[R]) exeFn() func() (R, error) {
	r.currentIndex++
	return func(idx int) func() (R, error) {
		return func() (R, error) { return r.exe(idx) }
	}(r.currentIndex - 1)
}

func (r *concurrentIterator[R]) exeNext() (R, error) {
	r.currentIndex++
	return r.exe(r.currentIndex - 1)
}

func (r *concurrentIterator[R]) exeFnAsAny() func() (any, error) {
	r.currentIndex++
	return func(idx int) func() (any, error) {
		return func() (any, error) { return r.exe(idx) }
	}(r.currentIndex - 1)
}

func (r *concurrentIterator[R]) exeNextAsAny() (any, error) {
	return r.exeNext()
}
