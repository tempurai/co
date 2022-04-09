package co

import (
	"sync"
	"sync/atomic"

	"github.com/tempura-shrimp/co/pool"
	co_sync "github.com/tempura-shrimp/co/sync"
)

type seqItem[R any] struct {
	seq uint64
	val R
}

type asyncRoundTrip[R any, E any, T seqItem[R]] struct {
	async AsyncSequenceable[T]
	it    Iterator[T]

	globalSeq uint64
	sourceCh  chan T

	interminCh   chan T
	executorPool *pool.DispatcherPool[E]
	workerMap    sync.Map
	callbackMap  sync.Map
	runOnce      sync.Once
}

func NewAsyncRoundTrip[R any, E any, T seqItem[R]](fn func(E)) *asyncRoundTrip[R, E, T] {
	ch := make(chan T)
	a := &asyncRoundTrip[R, E, T]{
		async:        FromChanBuffered(ch),
		sourceCh:     ch,
		interminCh:   make(chan T),
		executorPool: pool.NewDispatchPool[E](512),
	}
	a.it = a.async.iterator()
	a.executorPool.SetCallbackFn(a.receiveCallback)
	a.startListening()
	return a
}

func (a *asyncRoundTrip[R, E, T]) SendItem(val R, callbackFn func(E)) *asyncRoundTrip[R, E, T] {
	seq := atomic.AddUint64(&a.globalSeq, 1)
	a.sourceCh <- T(seqItem[R]{seq, val})
	a.callbackMap.Store(seq, callbackFn)
	return a
}

func (a *asyncRoundTrip[R, E, T]) Transform(fn func(AsyncSequenceable[T]) AsyncSequenceable[T]) *asyncRoundTrip[R, E, T] {
	a.async = fn(a.async)
	a.it = a.async.iterator()
	return a
}

func (a *asyncRoundTrip[R, E, T]) SetAsyncSequenceable(async AsyncSequenceable[T]) *asyncRoundTrip[R, E, T] {
	a.async = async
	a.it = a.async.iterator()
	return a
}

func (a *asyncRoundTrip[R, E, T]) startListening() {
	a.runOnce.Do(func() {
		co_sync.SafeGo(func() {
			for op, err := a.it.next(); op.valid; op, err = a.it.next() {
				if err != nil {
					continue
				}
				a.interminCh <- op.data
			}
		})
	})
}

func (a *asyncRoundTrip[R, E, T]) Done() *asyncRoundTrip[R, E, T] {
	co_sync.SafeClose(a.sourceCh)
	a.executorPool.Wait().Stop()
	co_sync.SafeClose(a.interminCh)
	return a
}

func (a *asyncRoundTrip[R, E, T]) Handle(fn func(R) E) {
	for item := range a.interminCh {
		func(item seqItem[R]) {
			pSeq := a.executorPool.ReserveSeq()
			a.workerMap.Store(pSeq, item.seq)
			a.executorPool.AddJobAt(pSeq, func() E {
				return fn(item.val)
			})
		}(seqItem[R](item))
	}
}

func (a *asyncRoundTrip[R, E, T]) receiveCallback(pSeq uint64, val E) {
	seq, ok := a.workerMap.Load(pSeq)
	if !ok {
		panic("co / roundTrip: pSeq to seq map not found")
	}
	callbackFn, ok := a.callbackMap.Load(seq.(uint64))
	if !ok {
		panic("co / roundTrip: seq to callback fn map not found")
	}
	co_sync.SafeGo(func() {
		callbackFn.(func(E))(val)
	})
}
