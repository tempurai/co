package co

import (
	"sync"
)

type zipBasic struct {
	combineLatestBasic

	updated      []int
	currentIndex int
}

func NewZipBasic(sequences []GenericExecutableSequence, fn func([]any, error, bool)) *zipBasic {
	return &zipBasic{
		combineLatestBasic: *NewCombineLatestBasic(sequences, fn),
		currentIndex:       1,
	}
}

type zipResult struct {
	combineLatestResult
}

func (z *zipBasic) ifReachesToIndexOrEnd(idx int) bool {
	for i := range z.sequences {
		if z.updated[i] != idx && !z.sequences[i].hasNext() {
			return false
		}
	}
	return true
}

func (z *zipBasic) exe() {
	resultChan := make(chan combineLatestResult)

	wg := sync.WaitGroup{}
	wg.Add(len(z.sequences))

	cond := sync.Cond{}
	for i := range z.sequences {
		wg.Add(1)

		go func(idx int, seq GenericExecutableSequence) {
			defer wg.Done()

			for i := 0; seq.hasNext(); i++ {
				cond.Wait()

				data, err := seq.exeNextAsAny()
				SafeSend(resultChan, combineLatestResult{idx, data, err})
			}
		}(i, z.sequences[i])
	}

	latestResults := make([]any, len(z.sequences))
	go func() {
		for {
			select {
			case result := <-resultChan:
				latestResults[result.index] = result.data
				z.updated[result.index]++

				if !EvertGET(z.updated, 1) {
					continue
				}
				if !z.ifReachesToIndexOrEnd(z.currentIndex) {
					continue
				}
				z.currentIndex++

				rte := z.ifAllSequenceReachesToEnd()
				z.fn(latestResults, result.err, rte)

				cond.Broadcast()
				if rte {
					return
				}
			}
		}
	}()

	wg.Wait()
}

func Zip[T1, T2 any](fn func(T1, T2, error, bool), seq1 Concurrently[T1], seq2 Concurrently[T2]) {
	NewZipBasic(castToGenericExecutableSequence(seq1.Iterator(), seq2.Iterator()), func(a []any, err error, b bool) {
		fn(CastOrNil[T1](a[0]), CastOrNil[T2](a[1]), err, b)
	}).exe()
}

func Zip3[T1, T2, T3 any](fn func(T1, T2, T3, error, bool), seq1 Concurrently[T1], seq2 Concurrently[T2], seq3 Concurrently[T3]) {
	NewZipBasic(castToGenericExecutableSequence(seq1.Iterator(), seq2.Iterator(), seq3.Iterator()), func(a []any, err error, b bool) {
		fn(CastOrNil[T1](a[0]), CastOrNil[T2](a[1]), CastOrNil[T3](a[2]), err, b)
	}).exe()
}
