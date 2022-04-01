package co

type AsyncExecutable[R any] struct {
	executables *executablesList[R]

	_defaultIterator *asyncExecutableIterator[R]
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

func (c *AsyncExecutable[R]) AddExecutable(fns ...func() (R, error)) *AsyncExecutable[R] {
	c.executables.AddExecutable(fns...)
	return c
}

func (c *AsyncExecutable[R]) defaultIterator() *asyncExecutableIterator[R] {
	if c._defaultIterator != nil {
		return c._defaultIterator
	}
	c._defaultIterator = c.Iterator()
	return c._defaultIterator
}

func (c *AsyncExecutable[R]) Iterator() *asyncExecutableIterator[R] {
	it := &asyncExecutableIterator[R]{
		AsyncExecutable: c,
		underlying:      c.executables.iterativeList.Iterator(),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncExecutableIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncExecutable[R]
	underlying *iterativeListIterator[*executable[R]]
}

func (it *asyncExecutableIterator[R]) preflight() bool {
	return it.underlying.preflight()
}

func (it *asyncExecutableIterator[R]) consume() (R, error) {
	defer func() { it.underlying.currentIndex++ }()
	return it.executeAt(it.underlying.currentIndex)
}
