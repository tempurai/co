package co

import (
	co_sync "go.tempura.ink/co/internal/sync"
)

type AsyncSubject[R any] struct {
	*asyncSequence[R]

	latestDataCh chan *data[R]
	sourceEnded  bool
}

func NewAsyncSubject[R any]() *AsyncSubject[R] {
	a := &AsyncSubject[R]{
		latestDataCh: make(chan *data[R]),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (c *AsyncSubject[R]) Next(val R) *AsyncSubject[R] {
	co_sync.SafeGo(func() {
		co_sync.SafeNRead(c.latestDataCh)
		c.latestDataCh <- NewDataWith(val, nil)
	})
	return c
}

func (c *AsyncSubject[R]) Error(err error) *AsyncSubject[R] {
	co_sync.SafeGo(func() {
		co_sync.SafeNRead(c.latestDataCh)
		c.latestDataCh <- NewDataWith(*new(R), err)
	})
	return c
}

func (c *AsyncSubject[R]) Complete() *AsyncSubject[R] {
	c.sourceEnded = true
	co_sync.SafeClose(c.latestDataCh)
	return c
}

func (c *AsyncSubject[R]) iterator() Iterator[R] {
	it := &asyncSubjectIterator[R]{
		AsyncSubject: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncSubjectIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncSubject[R]
}

func (it *asyncSubjectIterator[R]) next() *Optional[R] {
	if it.sourceEnded {
		return NewOptionalEmpty[R]()
	}

	for data, ok := <-it.latestDataCh; ok; data, ok = <-it.latestDataCh {
		if data.err != nil {
			it.handleError(data.GetError())
			if it.errorMode.shouldSkip() {
				continue
			}
			if it.errorMode.shouldStop() {
				return NewOptionalEmpty[R]()
			}
		}
		return OptionalOf(data.GetValue())
	}

	return NewOptionalEmpty[R]()
}
