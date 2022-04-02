package co

type AsyncExecutable[R any] struct {
	*asyncSequence[R]

	executables *executablesList[R]
}

func NewAsyncExecutable[R any]() *AsyncExecutable[R] {
	a := &AsyncExecutable[R]{
		executables: NewExecutablesList[R](),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
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

func (c *AsyncExecutable[R]) Iterator() Iterator[R] {
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

func (it *asyncExecutableIterator[R]) next() (*Optional[R], error) {
	if !it.underlying.preflight() {
		return NewOptionalEmpty[R](), nil
	}

	defer func() { it.underlying.currentIndex++ }()
	val, err := it.executeAt(it.underlying.currentIndex)
	return OptionalOf(val), err
}
