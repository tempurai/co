package co

type CoExecutableSequence[R any] struct {
	executables *executablesList[R]
	data        *determinedDataList[R]

	_defaultIterator Iterator[R]
}

func NewCoExecutableSequence[R any]() *CoExecutableSequence[R] {
	return &CoExecutableSequence[R]{
		executables: NewExecutablesList[R](),
		data:        NewDeterminedDataList[R](),
	}
}

func (c *CoExecutableSequence[R]) len() int {
	return c.executables.len()
}

func (c *CoExecutableSequence[R]) exe(i int) (R, error) {
	if c.executables.getAt(i).isExecuted() {
		return c.data.getAt(i)
	}

	return c.forceExeAt(i)
}

func (c *CoExecutableSequence[R]) forceExeAt(i int) (R, error) {
	val, err := c.executables.getAt(i).exe()
	c.data.setAt(i, val, err)

	return val, err
}

func (c *CoExecutableSequence[R]) AddFn(fns ...func() (R, error)) *CoExecutableSequence[R] {
	for i := range fns {
		c.data.List.add(NewData[R]())

		e := NewExecutor[R]()
		e.fn = fns[i]
		c.executables.add(e)
	}
	return c
}

func (c *CoExecutableSequence[R]) defaultIterator() Iterator[R] {
	if c._defaultIterator != nil {
		return c._defaultIterator
	}
	c._defaultIterator = c.Iterator()
	return c._defaultIterator
}

func (c *CoExecutableSequence[R]) Iterator() Iterator[R] {
	return &coExecutableSequenceIterator[R]{
		CoExecutableSequence:  c,
		iterativeListIterator: c.executables.iterativeList.Iterator(),
	}
}

type coExecutableSequenceIterator[R any] struct {
	*CoExecutableSequence[R]
	iterativeListIterator[*executable[R]]
}

func (it *coExecutableSequenceIterator[R]) next() (R, error) {
	it.currentIndex++
	return it.exe(it.currentIndex)
}

func (it *coExecutableSequenceIterator[R]) dispatch() func() (R, error) {
	fn := Copy(it).next
	it.currentIndex++
	return fn
}

func (it *coExecutableSequenceIterator[R]) nextAny() (any, error) {
	return it.next()
}
