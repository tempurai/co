package co

import (
	"time"

	syncx "go.tempura.ink/co/internal/sync"
)

type AsyncInterval[R any] struct {
	*asyncSequence[R]

	period int
	ended  syncx.AtomicBool
}

func Interval(period int) *AsyncInterval[int] {
	a := &AsyncInterval[int]{
		period: period,
	}
	a.asyncSequence = NewAsyncSequence[int](a)
	return a
}

func (a *AsyncInterval[R]) Complete() *AsyncInterval[R] {
	a.ended.Set(true)
	return a
}

func (a *AsyncInterval[R]) iterator() Iterator[R] {
	it := &asyncIntervalIterator[R]{
		AsyncInterval: a,
		timer:         time.NewTicker(time.Duration(a.period) * time.Millisecond),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncIntervalIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncInterval[R]

	timer *time.Ticker
}

func (it *asyncIntervalIterator[R]) cleanUp() *Optional[R] {
	it.timer.Stop()
	return NewOptionalEmpty[R]()
}

func (it *asyncIntervalIterator[R]) next() *Optional[R] {
	if it.ended.Get() {
		return it.cleanUp()
	}

	<-it.timer.C
	return OptionalOf(*new(R))
}
