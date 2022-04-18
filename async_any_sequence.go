package co

import (
	"sync"

	"tempura.ink/co/ds/queue"
	co_sync "tempura.ink/co/internal/sync"
)

type AsyncAnySequence[R any] struct {
	*asyncSequence[R]

	its []Iterator[R]
}

func NewAsyncAnySequence[R any](its ...AsyncSequenceable[R]) *AsyncAnySequence[R] {
	a := &AsyncAnySequence[R]{
		its: toAsyncIterators(its...),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (c *AsyncAnySequence[R]) iterator() Iterator[R] {
	it := &asyncAnySequenceIterator[R]{
		AsyncAnySequence: c,
		dataQueue:        queue.NewQueue[R](),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncAnySequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncAnySequence[R]
	dataQueue   *queue.Queue[R]
	sourceEnded bool
}

func (it *asyncAnySequenceIterator[R]) next() *Optional[R] {
	if it.sourceEnded {
		return NewOptionalEmpty[R]()
	}
	if it.dataQueue.Len() != 0 {
		return OptionalOf(it.dataQueue.Dequeue())
	}

	completedCh := make(chan bool)

	wg := &sync.WaitGroup{}
	wg.Add(len(it.its))

	for i, pIt := range it.its {
		go func(idx int, pIt Iterator[R]) {
			defer wg.Done()
			for op := pIt.next(); op.valid; op = pIt.next() {
				it.dataQueue.Enqueue(op.data)
				co_sync.SafeSend(completedCh, true)
				return
			}
			co_sync.SafeSend(completedCh, true)
		}(i, pIt)
	}

	<-completedCh
	close(completedCh)

	if it.dataQueue.Len() > 0 {
		return OptionalOf(it.dataQueue.Dequeue())
	}

	if wg.Wait(); it.dataQueue.Len() == 0 {
		it.sourceEnded = true
		return NewOptionalEmpty[R]()
	}
	return OptionalOf(it.dataQueue.Dequeue())
}
