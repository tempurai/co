package co

import (
	"sync"
)

type actionCombineLatest[R any] struct {
	*Action[R]

	its []IteratorAny
	fn  func(*actionCombineLatest[R], []any, error, bool)
}

func NewActionCombineLatest[R any](iterators []IteratorAny) *actionCombineLatest[R] {
	return &actionCombineLatest[R]{
		Action: NewAction[R](),
		its:    iterators,
	}
}

func (a *actionCombineLatest[R]) setFn(fn func(*actionCombineLatest[R], []any, error, bool)) *actionCombineLatest[R] {
	a.fn = fn
	return a
}

func (a *actionCombineLatest[R]) ifAllSequenceReachesToEnd() bool {
	for _, it := range a.its {
		if it.hasNext() {
			return false
		}
	}
	return true
}

func (a *actionCombineLatest[R]) run() {
	resultChan := make(chan actionAnyResult)

	wg := sync.WaitGroup{}
	wg.Add(len(a.its))

	for i := range a.its {
		wg.Add(1)

		go func(idx int, seq IteratorAny) {
			defer wg.Done()

			for seq.hasNext() {
				data, err := seq.nextAny()
				SafeSend(resultChan, actionAnyResult{idx, data, err})
			}
		}(i, a.its[i])
	}

	latestResults := make([]any, len(a.its))
	updated := make([]int, len(a.its))
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

				rte := a.ifAllSequenceReachesToEnd()
				a.fn(a, latestResults, result.err, rte)

				if rte {
					return
				}
			}
		}
	}()

	wg.Wait()
	a.done()
}

func CombineLatest[T1, T2 any](seq1 CoSequenceable[T1], seq2 CoSequenceable[T2]) *Action[ActionBulkResult[Type2[T1, T2]]] {
	action := NewActionCombineLatest[ActionBulkResult[Type2[T1, T2]]](castToIteratorAny(seq1.Iterator(), seq2.Iterator())).
		setFn(func(a *actionCombineLatest[ActionBulkResult[Type2[T1, T2]]], v []any, err error, b bool) {
			a.listenProgressive(ActionBulkResult[Type2[T1, T2]]{
				Value:        Type2[T1, T2]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1])},
				Err:          err,
				ReachesToEnd: b,
			})
		})

	SafeGo(action.run)
	return action.Action
}

func CombineLatest3[T1, T2, T3 any](seq1 CoSequenceable[T1], seq2 CoSequenceable[T2], seq3 CoSequenceable[T3]) *Action[ActionBulkResult[Type3[T1, T2, T3]]] {
	action := NewActionCombineLatest[ActionBulkResult[Type3[T1, T2, T3]]](castToIteratorAny(seq1.Iterator(), seq2.Iterator())).
		setFn(func(a *actionCombineLatest[ActionBulkResult[Type3[T1, T2, T3]]], v []any, err error, b bool) {
			a.listenProgressive(ActionBulkResult[Type3[T1, T2, T3]]{
				Value:        Type3[T1, T2, T3]{CastOrNil[T1](v[0]), CastOrNil[T2](v[1]), CastOrNil[T3](v[2])},
				Err:          err,
				ReachesToEnd: b,
			})
		})

	SafeGo(action.run)
	return action.Action
}
