package co

import (
	"sync"
)

type combineLatestBasic struct {
	sequences []GenericExecutableSequence
	fn        func([]any, error, bool)
}

func NewCombineLatestBasic(sequences []GenericExecutableSequence, fn func([]any, error, bool)) *combineLatestBasic {
	return &combineLatestBasic{
		sequences: sequences,
		fn:        fn,
	}
}

type combineLatestResult struct {
	index int
	data  any
	err   error
}

func (c *combineLatestBasic) ifAllSequenceReachesToEnd() bool {
	for _, seq := range c.sequences {
		if seq.hasNext() {
			return false
		}
	}
	return true
}

func (c *combineLatestBasic) exe() {
	resultChan := make(chan combineLatestResult)

	wg := sync.WaitGroup{}
	wg.Add(len(c.sequences))

	for i := range c.sequences {
		wg.Add(1)

		go func(idx int, seq GenericExecutableSequence) {
			defer wg.Done()

			for i := 0; seq.hasNext(); i++ {
				data, err := seq.exeNextAsAny()
				SafeSend(resultChan, combineLatestResult{idx, data, err})
			}
		}(i, c.sequences[i])
	}

	latestResults := make([]any, len(c.sequences))
	updated := make([]int, len(c.sequences))
	arrayFilled := false

	go func() {
		for {
			select {
			case result := <-resultChan:
				latestResults[result.index] = result.data
				updated[result.index]++

				if !arrayFilled && !EvertGET(updated, 1) {
					continue
				}
				arrayFilled = true

				rte := c.ifAllSequenceReachesToEnd()
				c.fn(latestResults, result.err, rte)

				if rte {
					return
				}
			}
		}
	}()

	wg.Wait()
}

func CombineLatest[T1, T2 any](fn func(T1, T2, error, bool), seq1 Concurrently[T1], seq2 Concurrently[T2]) {
	NewCombineLatestBasic(castToGenericExecutableSequence(seq1.Iterator(), seq2.Iterator()), func(a []any, err error, b bool) {
		fn(CastOrNil[T1](a[0]), CastOrNil[T2](a[1]), err, b)
	}).exe()
}

func CombineLatest3[T1, T2, T3 any](fn func(T1, T2, T3, error, bool), seq1 Concurrently[T1], seq2 Concurrently[T2], seq3 Concurrently[T3]) {
	NewCombineLatestBasic(castToGenericExecutableSequence(seq1.Iterator(), seq2.Iterator(), seq3.Iterator()), func(a []any, err error, b bool) {
		fn(CastOrNil[T1](a[0]), CastOrNil[T2](a[1]), CastOrNil[T3](a[2]), err, b)
	}).exe()
}
