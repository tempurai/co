package co

import (
	"sync"
)

type asyncCombineLatestFn[R any] func([]any, error) (R, error)

type AsyncCombineLatestSequence[R any] struct {
	its         []iteratorAny
	converterFn asyncCombineLatestFn[R]
}

func NewAsyncCombineLatestSequence[R any](its []iteratorAny) *AsyncCombineLatestSequence[R] {
	return &AsyncCombineLatestSequence[R]{
		its: its,
	}
}

func (a *AsyncCombineLatestSequence[R]) setConverterFn(fn asyncCombineLatestFn[R]) *AsyncCombineLatestSequence[R] {
	a.converterFn = fn
	return a
}

func (a *AsyncCombineLatestSequence[R]) iterator() Iterator[R] {
	it := &asyncCombineLatestSequenceIterator[R]{
		AsyncCombineLatestSequence: a,
		dataStore:                  make([]any, len(a.its)),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

func CombineLatest[T1, T2 any](seq1 AsyncSequenceable[T1], seq2 AsyncSequenceable[T2]) *AsyncCombineLatestSequence[Type2[T1, T2]] {
	anyIterators := castToIteratorAny(seq1.iterator(), seq2.iterator())
	converterFn := func(v []any, err error) (Type2[T1, T2], error) {
		return Type2[T1, T2]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1])}, err
	}

	a := NewAsyncCombineLatestSequence[Type2[T1, T2]](anyIterators).setConverterFn(converterFn)
	return a
}

func CombineLatest3[T1, T2, T3 any](seq1 AsyncSequenceable[T1], seq2 AsyncSequenceable[T2], seq3 AsyncSequenceable[T3]) *AsyncCombineLatestSequence[Type3[T1, T2, T3]] {
	anyIterators := castToIteratorAny(seq1.iterator(), seq2.iterator())
	converterFn := func(v []any, err error) (Type3[T1, T2, T3], error) {
		return Type3[T1, T2, T3]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1]), CastOrNil[T3](v[2])}, err
	}
	a := NewAsyncCombineLatestSequence[Type3[T1, T2, T3]](anyIterators).setConverterFn(converterFn)
	return a
}

type asyncCombineLatestSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncCombineLatestSequence[R]
	dataStore             []any
	firstpassCompleted    bool
	normalPassCompletedCh chan bool
	updated               []bool
}

func (a *asyncCombineLatestSequenceIterator[R]) pass() {
	wg := sync.WaitGroup{}
	wg.Add(len(a.its))

	a.updated = make([]bool, len(a.its))
	for i, it := range a.its {
		go func(idx int, it iteratorAny) {
			defer wg.Done()
			for op, err := it.nextAny(); op.valid && err != nil; op, err = it.nextAny() {
				a.dataStore[idx] = op.data
				if !a.firstpassCompleted {
					a.normalPassCompletedCh <- true
				}
				a.firstpassCompleted = true
				a.updated[idx] = true
			}
		}(i, it)
	}

	wg.Wait()
}

func (a *asyncCombineLatestSequenceIterator[R]) next() (*Optional[R], error) {
	if !a.firstpassCompleted {
		SafeGo(a.pass)
		<-a.normalPassCompletedCh
	} else {
		a.pass()
	}

	if EvertET(a.updated, false) {
		return NewOptionalEmpty[R](), nil
	}

	data, err := a.converterFn(a.dataStore, nil)
	return OptionalOf(data), err
}
