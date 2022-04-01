package co

type AsyncExecutable[R any] struct {
	executables *executablesList[R]
	data        *determinedDataList[R]

	_defaultIterator Iterator[R]
}

func NewAsyncExecutable[R any]() *AsyncExecutable[R] {
	return &AsyncExecutable[R]{
		executables: NewExecutablesList[R](),
		data:        NewDeterminedDataList[R](),
	}
}

func (c *AsyncExecutable[R]) len() int {
	return c.executables.len()
}

func (c *AsyncExecutable[R]) exe(i int) (R, error) {
	if c.executables.getAt(i).isExecuted() {
		return c.data.getAt(i)
	}

	return c.forceExeAt(i)
}

func (c *AsyncExecutable[R]) forceExeAt(i int) (R, error) {
	val, err := c.executables.getAt(i).exe()
	c.data.setAt(i, val, err)

	return val, err
}

func (c *AsyncExecutable[R]) AddFn(fns ...func() (R, error)) *AsyncExecutable[R] {
	for i := range fns {
		c.data.List.add(NewData[R]())

		e := NewExecutor[R]()
		e.fn = fns[i]
		c.executables.add(e)
	}
	return c
}

func (c *AsyncExecutable[R]) defaultIterator() Iterator[R] {
	if c._defaultIterator != nil {
		return c._defaultIterator
	}
	c._defaultIterator = c.Iterator()
	return c._defaultIterator
}

func (c *AsyncExecutable[R]) Iterator() Iterator[R] {
	return &asyncExecutableIterator[R]{
		AsyncExecutable:       c,
		iterativeListIterator: c.executables.iterativeList.Iterator(),
	}
}

type asyncExecutableIterator[R any] struct {
	*AsyncExecutable[R]
	iterativeListIterator[*executable[R]]
}

func (it *asyncExecutableIterator[R]) next() (R, error) {
	defer func() { it.currentIndex++ }()
	return it.exe(it.currentIndex)
}
