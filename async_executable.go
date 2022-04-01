package co

type AsyncExecutable[R any] struct {
	executables *executablesList[R]

	_defaultIterator Iterator[R]
}

func NewAsyncExecutable[R any]() *AsyncExecutable[R] {
	return &AsyncExecutable[R]{
		executables: NewExecutablesList[R](),
	}
}

func (c *AsyncExecutable[R]) len() int {
	return c.executables.len()
}

func (c *AsyncExecutable[R]) executeAt(i int) (R, error) {
	val, err := c.executables.getAt(i).exe()
	return val, err
}

func (c *AsyncExecutable[R]) AddFn(fns ...func() (R, error)) *AsyncExecutable[R] {
	c.executables.AddFn(fns...)
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
	return it.executeAt(it.currentIndex)
}
