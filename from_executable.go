package co

import "sync/atomic"

type AsyncExecutable[R any] struct {
	*asyncSequence[R]

	executables *executablesList[R]
}

func FromExecutable[R any]() *AsyncExecutable[R] {
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

func (c *AsyncExecutable[R]) iterator() Iterator[R] {
	it := &asyncExecutableIterator[R]{
		AsyncExecutable: c,
		underlying:      c.executables.iterativeList.iterator(),
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

	defer func() { atomic.AddInt32(&it.underlying.currentIndex, 1) }()
	val, err := it.executeAt(int(it.underlying.currentIndex))
	return OptionalOf(val), err
}
