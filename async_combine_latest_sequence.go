package co

import (
	"sync"

	"github.com/tempurai/co/ds/queue"
	syncx "github.com/tempurai/co/internal/syncx"
)

type asyncCombineLatestFn[R any] func([]any) R

type AsyncCombineLatestSequence[R any] struct {
	*asyncSequence[R]

	its         []iteratorAny
	converterFn asyncCombineLatestFn[R]
}

func NewAsyncCombineLatestSequence[R any](its []iteratorAny) *AsyncCombineLatestSequence[R] {
	a := &AsyncCombineLatestSequence[R]{
		its: its,
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (a *AsyncCombineLatestSequence[R]) setConverterFn(fn asyncCombineLatestFn[R]) *AsyncCombineLatestSequence[R] {
	a.converterFn = fn
	return a
}

func (a *AsyncCombineLatestSequence[R]) iterator() Iterator[R] {
	it := &asyncCombineLatestSequenceIterator[R]{
		AsyncCombineLatestSequence: a,
		dataStore:                  make([]any, len(a.its)),
		statusData:                 make([]combineLatestAsyncData, len(a.its)),
		bufferWait:                 syncx.NewCondx(&sync.Mutex{}),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	for i := range it.statusData {
		it.statusData[i] = combineLatestAsyncData{
			asyncData: *NewAsyncData(queue.NewQueue[any]()).TransiteTo(asyncStatusDone),
		}
	}
	return it
}

func CombineLatest[T1, T2 any](seq1 AsyncSequenceable[T1], seq2 AsyncSequenceable[T2]) *AsyncCombineLatestSequence[Type2[T1, T2]] {
	anyIterators := castToIteratorAny(seq1.iterator(), seq2.iterator())
	converterFn := func(v []any) Type2[T1, T2] {
		return Type2[T1, T2]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1])}
	}

	a := NewAsyncCombineLatestSequence[Type2[T1, T2]](anyIterators).setConverterFn(converterFn)
	return a
}

func CombineLatest3[T1, T2, T3 any](seq1 AsyncSequenceable[T1], seq2 AsyncSequenceable[T2], seq3 AsyncSequenceable[T3]) *AsyncCombineLatestSequence[Type3[T1, T2, T3]] {
	anyIterators := castToIteratorAny(seq1.iterator(), seq2.iterator(), seq3.iterator())
	converterFn := func(v []any) Type3[T1, T2, T3] {
		return Type3[T1, T2, T3]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1]), CastOrNil[T3](v[2])}
	}
	a := NewAsyncCombineLatestSequence[Type3[T1, T2, T3]](anyIterators).setConverterFn(converterFn)
	return a
}

type combineLatestAsyncData struct {
	asyncData   AsyncData[*queue.Queue[any]]
	sourceEnded bool
}

type asyncCombineLatestSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncCombineLatestSequence[R]
	firstpass  bool
	bufferWait *syncx.Condx

	statusData []combineLatestAsyncData
	dataStore  []any

	mux sync.Mutex // Avoid data race
}

func (a *asyncCombineLatestSequenceIterator[R]) copyDirtyToStore() {
	a.mux.Lock()
	defer a.mux.Unlock()

	for i := range a.statusData {
		if !a.statusData[i].asyncData.IsComplete() {
			continue
		}

		a.dataStore[i] = a.statusData[i].asyncData.value.Dequeue()
		if a.statusData[i].asyncData.value.Len() == 0 && a.statusData[i].asyncData.IsComplete() {
			a.statusData[i].asyncData.TransiteTo(asyncStatusIdle)
		}
	}
}

func (a *asyncCombineLatestSequenceIterator[R]) hasQueuedData() bool {
	a.mux.Lock()
	defer a.mux.Unlock()

	for i := range a.statusData {
		if a.statusData[i].asyncData.value.Len() != 0 {
			return true
		}
	}
	return false
}

func (a *asyncCombineLatestSequenceIterator[R]) allSourcesEneded() bool {
	a.mux.Lock()
	defer a.mux.Unlock()

	for i := range a.statusData {
		if !a.statusData[i].sourceEnded {
			return false
		}
	}
	return true
}

func (a *asyncCombineLatestSequenceIterator[R]) startFirstPass() {
	wg := sync.WaitGroup{}
	wg.Add(len(a.its))

	for i, it := range a.its {
		go func(idx int, it iteratorAny) {
			defer wg.Done()

			for op := it.nextAny(); op.valid; op = it.nextAny() {
				a.statusData[idx].asyncData.value.Enqueue(op.data)
				break
			}
		}(i, it)
	}

	wg.Wait()
	a.firstpass = true
}

func (a *asyncCombineLatestSequenceIterator[R]) pass() {
	if !a.firstpass {
		a.startFirstPass()
		a.bufferWait.Broadcast()
		return
	}

	for i := range a.its {
		a.mux.Lock()
		sourceEnded := a.statusData[i].sourceEnded
		a.mux.Unlock()
		if sourceEnded {
			continue
		}
		if !a.statusData[i].asyncData.IsIdel() {
			continue
		}
		if a.statusData[i].asyncData.value.Len() > 0 {
			continue
		}
		a.statusData[i].asyncData.TransiteTo(asyncStatusPending)

		go func(idx int, statusData *combineLatestAsyncData, it iteratorAny) {
			for op := it.nextAny(); op.valid; op = it.nextAny() {
				a.bufferWait.Broadcastify(&syncx.BroadcastOption{
					PreProcessFn: func() {
						a.mux.Lock()
						defer a.mux.Unlock()

						statusData.asyncData.value.Enqueue(op.data)
						statusData.asyncData.TransiteTo(asyncStatusDone)
					}},
				)
				return
			}
			a.bufferWait.Broadcastify(&syncx.BroadcastOption{
				PreProcessFn: func() {
					a.mux.Lock()
					defer a.mux.Unlock()

					// becuase we have already emmited final value at loop therefore no more data
					statusData.asyncData.TransiteTo(asyncStatusIdle)
					statusData.sourceEnded = true
				}},
			)
		}(i, &a.statusData[i], a.its[i])
	}
}

func (a *asyncCombineLatestSequenceIterator[R]) next() *Optional[R] {
	if !a.allSourcesEneded() && !a.hasQueuedData() {
		a.pass()
		a.bufferWait.Waitify(&syncx.WaitOption{
			ConditionFn: func() bool {
				return !a.allSourcesEneded() && !a.hasQueuedData()
			},
		})
	}
	if a.allSourcesEneded() && !a.hasQueuedData() {
		return NewOptionalEmpty[R]()
	}

	a.copyDirtyToStore()
	data := a.converterFn(a.dataStore)
	return OptionalOf(data)
}
