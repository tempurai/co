package co

import (
	"log"
	"sync"

	"github.com/tempurai/co/ds/queue"
	syncx "github.com/tempurai/co/internal/syncx"
)

type AsyncMulticastSequence[R any] struct {
	previousIterator Iterator[R]
	sourceEnded      bool
	runOnce          sync.Once

	bufferedQueue *queue.MultiReceiverQueue[R]
	bufferWait    *syncx.Condx
}

func NewAsyncMulticastSequence[R any](it AsyncSequenceable[R]) *AsyncMulticastSequence[R] {
	a := &AsyncMulticastSequence[R]{
		previousIterator: it.iterator(),
		bufferedQueue:    queue.NewMultiReceiverQueue[R](),
		bufferWait:       syncx.NewCondx(&sync.Mutex{}),
	}
	a.startListening()
	return a
}

func (a *AsyncMulticastSequence[T]) startListening() {
	a.runOnce.Do(func() {
		syncx.SafeGo(func() {
			for op := a.previousIterator.next(); op.valid; op = a.previousIterator.next() {
				a.bufferWait.Broadcastify(&syncx.BroadcastOption{
					PreProcessFn: func() {
						a.bufferedQueue.Enqueue(op.data)
					},
				})
			}
			a.bufferWait.Broadcastify(&syncx.BroadcastOption{
				PreProcessFn: func() { a.sourceEnded = true },
			})
		})
	})
}

func (a *AsyncMulticastSequence[R]) Connect() *AsyncMulticastConnector[R] {
	c := &AsyncMulticastConnector[R]{
		AsyncMulticastSequence: a,
	}
	c.asyncSequence = NewAsyncSequence[R](c)
	c._defaultIterator = c.iterator() // init here to ensure queue works
	return c
}

type AsyncMulticastConnector[R any] struct {
	*asyncSequence[R]

	*AsyncMulticastSequence[R]
}

func (a *AsyncMulticastConnector[R]) iterator() Iterator[R] {
	it := &asyncMulticastSequenceIterator[R]{
		AsyncMulticastSequence: a.AsyncMulticastSequence,
		receiver:               a.bufferedQueue.Receiver(),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncMulticastSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncMulticastSequence[R]
	receiver *queue.QueueReceiver[R]
}

func (it *asyncMulticastSequenceIterator[R]) next() *Optional[R] {
	it.bufferWait.Waitify(&syncx.WaitOption{
		ConditionFn: func() bool {
			return !it.sourceEnded && it.receiver.IsEmpty()
		},
	})

	log.Println("it.sourceEnded && it.receiver.IsEmpty() ", it.sourceEnded, it.receiver.IsEmpty())
	if it.sourceEnded && it.receiver.IsEmpty() {
		return NewOptionalEmpty[R]()
	}

	return OptionalOf(it.receiver.Dequeue())
}
