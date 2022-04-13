package co

import (
	"sync"
)

type asyncZipFn[R any] func([]any) R

type AsyncZipSequence[R any] struct {
	*asyncSequence[R]

	its         []iteratorAny
	converterFn asyncZipFn[R]
}

func NewAsyncZipSequence[R any](its []iteratorAny) *AsyncZipSequence[R] {
	a := &AsyncZipSequence[R]{
		its: its,
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (a *AsyncZipSequence[R]) setConverterFn(fn asyncZipFn[R]) *AsyncZipSequence[R] {
	a.converterFn = fn
	return a
}

func (a *AsyncZipSequence[R]) iterator() Iterator[R] {
	it := &asyncZipSequenceIterator[R]{
		AsyncZipSequence: a,
		latestData:       make([]any, len(a.its)),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

func Zip[T1, T2 any](seq1 AsyncSequenceable[T1], seq2 AsyncSequenceable[T2]) *AsyncZipSequence[Type2[T1, T2]] {
	anyIterators := castToIteratorAny(seq1.iterator(), seq2.iterator())
	converterFn := func(v []any) Type2[T1, T2] {
		return Type2[T1, T2]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1])}
	}

	a := NewAsyncZipSequence[Type2[T1, T2]](anyIterators).setConverterFn(converterFn)
	return a
}

func Zip3[T1, T2, T3 any](seq1 AsyncSequenceable[T1], seq2 AsyncSequenceable[T2], seq3 AsyncSequenceable[T3]) *AsyncZipSequence[Type3[T1, T2, T3]] {
	anyIterators := castToIteratorAny(seq1.iterator(), seq2.iterator(), seq3.iterator())
	converterFn := func(v []any) Type3[T1, T2, T3] {
		return Type3[T1, T2, T3]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1]), CastOrNil[T3](v[2])}
	}
	a := NewAsyncZipSequence[Type3[T1, T2, T3]](anyIterators).setConverterFn(converterFn)
	return a
}

type asyncZipSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncZipSequence[R]
	latestData []any
}

func (a *asyncZipSequenceIterator[R]) copyToLatest(waitedData []Optional[any]) bool {
	ifUpdated := false
	for i := range waitedData {
		if waitedData[i].valid {
			ifUpdated = true
			a.latestData[i] = waitedData[i].data
		}
	}
	return ifUpdated
}

func (a *asyncZipSequenceIterator[R]) next() *Optional[R] {
	waitedData := make([]Optional[any], len(a.its))

	wg := sync.WaitGroup{}
	wg.Add(len(a.its))

	for i, it := range a.its {
		go func(idx int, it iteratorAny) {
			defer wg.Done()
			for op := it.nextAny(); op.valid; op = it.nextAny() {
				waitedData[idx] = *OptionalOf(op.data)
				break
			}
		}(i, it)
	}
	wg.Wait()

	if !a.copyToLatest(waitedData) {
		return NewOptionalEmpty[R]()
	}

	data := a.converterFn(a.latestData)
	return OptionalOf(data)
}
