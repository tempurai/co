package co

import (
	"sync"
)

type asyncZipFn[R any] func([]any, error) (R, error)

type AsyncZipSequence[R any] struct {
	its         []IteratorAny
	converterFn asyncZipFn[R]
}

func NewAsyncZipSequence[R any](its []IteratorAny) *AsyncZipSequence[R] {
	return &AsyncZipSequence[R]{
		its: its,
	}
}

func (a *AsyncZipSequence[R]) setConverterFn(fn asyncZipFn[R]) *AsyncZipSequence[R] {
	a.converterFn = fn
	return a
}

func (a *AsyncZipSequence[R]) Iterator() Iterator[R] {
	it := &asyncZipSequenceIterator[R]{
		AsyncZipSequence: a,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

func Zip[T1, T2 any](seq1 AsyncSequenceable[T1], seq2 AsyncSequenceable[T2]) *AsyncZipSequence[Type2[T1, T2]] {
	anyIterators := castToIteratorAny(seq1.Iterator(), seq2.Iterator())
	converterFn := func(v []any, err error) (Type2[T1, T2], error) {
		return Type2[T1, T2]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1])}, err
	}

	a := NewAsyncZipSequence[Type2[T1, T2]](anyIterators).setConverterFn(converterFn)
	return a
}

func Zip3[T1, T2, T3 any](seq1 AsyncSequenceable[T1], seq2 AsyncSequenceable[T2], seq3 AsyncSequenceable[T3]) *AsyncZipSequence[Type3[T1, T2, T3]] {
	anyIterators := castToIteratorAny(seq1.Iterator(), seq2.Iterator())
	converterFn := func(v []any, err error) (Type3[T1, T2, T3], error) {
		return Type3[T1, T2, T3]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1]), CastOrNil[T3](v[2])}, err
	}
	a := NewAsyncZipSequence[Type3[T1, T2, T3]](anyIterators).setConverterFn(converterFn)
	return a
}

type asyncZipSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	*AsyncZipSequence[R]
}

func (a *asyncZipSequenceIterator[R]) next() (*Optional[R], error) {
	results := make([]any, len(a.its))
	updated := make([]bool, len(a.its))

	wg := sync.WaitGroup{}
	wg.Add(len(a.its))

	for i, it := range a.its {
		go func(idx int, it IteratorAny) {
			defer wg.Done()
			for op, err := it.nextAny(); op.valid && err != nil; {
				results[idx] = op.data
				updated[idx] = true
				break
			}
		}(i, it)
	}

	wg.Wait()

	if EvertET(updated, false) {
		return NewOptionalEmpty[R](), nil
	}

	data, err := a.converterFn(results, nil)
	return OptionalOf(data), err
}
