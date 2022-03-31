package co

type Concurrent[R any] struct {
	executables *executablesList[R]
	data        *determinedDataList[R]

	_defaultIterator ExecutableIterator[R]
}

func NewConcurrent[R any]() *Concurrent[R] {
	return &Concurrent[R]{
		executables: NewExecutablesList[R](),
		data:        NewDeterminedDataList[R](),
	}
}

func (c *Concurrent[R]) len() int {
	return c.executables.len()
}

func (c *Concurrent[R]) exe(i int) (R, error) {
	if c.executables.getAt(i).isExecuted() {
		return c.data.getAt(i)
	}

	return c.forceExeAt(i)
}

func (c *Concurrent[R]) forceExeAt(i int) (R, error) {
	val, err := c.executables.getAt(i).exe()
	c.data.setAt(i, val, err)

	return val, err
}

func (c *Concurrent[R]) addData(dVal ...*data[R]) *Concurrent[R] {
	c.data.List.add(dVal...)

	for range dVal {
		c.executables.add(NewExecutor[R]())
	}
	return c

}

func (c *Concurrent[R]) addExeFn(fns ...func() (R, error)) *Concurrent[R] {
	for i := range fns {
		c.data.List.add(NewData[R]())

		e := NewExecutor[R]()
		e.fn = fns[i]
		c.executables.add(e)
	}
	return c
}

func (c *Concurrent[R]) add(eVal ...*executable[R]) *Concurrent[R] {
	for range eVal {
		c.data.List.add(NewData[R]())
	}

	c.executables.add(eVal...)
	return c
}

func (c *Concurrent[R]) swapValue(dVal *determinedDataList[R], eVal *executablesList[R]) *Concurrent[R] {
	if dVal != nil {
		c.data.swap(dVal.list)
	}
	if eVal != nil {
		c.executables.swap(eVal.list)
	}
	return c
}

func (c *Concurrent[R]) defaultIterator() ExecutableIterator[R] {
	if c._defaultIterator != nil {
		return c._defaultIterator
	}
	c._defaultIterator = c.Iterator()
	return c._defaultIterator
}

func (c *Concurrent[R]) Iterator() ExecutableIterator[R] {
	return &concurrentIterator[R]{
		Concurrent:            c,
		iterativeListIterator: c.executables.iterativeList.Iterator(),
	}
}

type concurrentIterator[R any] struct {
	*Concurrent[R]
	*iterativeListIterator[*executable[R]]
}

func (it *concurrentIterator[R]) exeFn() func() (R, error) {
	it.currentIndex++
	return func(idx int) func() (R, error) {
		return func() (R, error) { return it.exe(idx) }
	}(it.currentIndex - 1)
}

func (it *concurrentIterator[R]) exeNext() (R, error) {
	it.currentIndex++
	return it.exe(it.currentIndex - 1)
}

func (it *concurrentIterator[R]) exeFnAsAny() func() (any, error) {
	it.currentIndex++
	return func(idx int) func() (any, error) {
		return func() (any, error) { return it.exe(idx) }
	}(it.currentIndex - 1)
}

func (it *concurrentIterator[R]) exeNextAsAny() (any, error) {
	return it.exeNext()
}
