package co

import (
	"sync"

	"go.tempura.ink/co/ds/queue"
	syncx "go.tempura.ink/co/internal/sync"
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
		completedCh:      make(chan bool),
		waitCond:         sync.NewCond(&sync.Mutex{}),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncAnySequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncAnySequence[R]
	dataQueue    *queue.Queue[R]
	sourceEnded  bool
	waitCond     *sync.Cond
	completedCh  chan bool
	ifProcessing syncx.AtomicBool
}

func (it *asyncAnySequenceIterator[R]) next() *Optional[R] {
	if it.sourceEnded {
		return NewOptionalEmpty[R]()
	}
	if it.dataQueue.Len() != 0 {
		return OptionalOf(it.dataQueue.Dequeue())
	}

	if it.ifProcessing.Get() {
		<-it.completedCh
	}
	it.ifProcessing.Set(true)

	wg := &sync.WaitGroup{}
	wg.Add(len(it.its))

	go func() {
		wg.Wait()
		it.ifProcessing.Set(false)
		syncx.SafeNSend(it.completedCh, true)

		if it.dataQueue.Len() > 0 {
			return
		}
		syncx.CondBroadcast(it.waitCond, func() { it.sourceEnded = true })
	}()

	for i, pIt := range it.its {
		go func(idx int, pIt Iterator[R]) {
			defer wg.Done()
			for op := pIt.next(); op.valid; op = pIt.next() {
				syncx.CondSignal(it.waitCond, func() { it.dataQueue.Enqueue(op.data) })
				return
			}
		}(i, pIt)
	}

	syncx.CondWait(it.waitCond, func() bool { return !it.sourceEnded && it.dataQueue.Len() == 0 })

	if it.sourceEnded {
		return NewOptionalEmpty[R]()
	}
	return OptionalOf(it.dataQueue.Dequeue())
}
