package co

import (
	"fmt"
)

type ConcurrentMulti[R any] struct {
	concurrents []Concurrent[R]
}

func NewConcurrentMulti[R any](cos ...Concurrent[R]) *ConcurrentMulti[R] {
	return &ConcurrentMulti[R]{
		concurrents: cos,
	}
}

func (r *ConcurrentMulti[R]) Iterator() *concurrentMultiIterator[R] {
	it := &concurrentMultiIterator[R]{}
	for i := range r.concurrents {
		it.iterators = append(it.iterators, r.concurrents[i].defaultIterator())
	}
	return it
}

type concurrentMultiIterator[R any] struct {
	iterators []ExecutableIterator[R]

	currentIndex int
}

func (r *concurrentMultiIterator[R]) nextIndex() int {
	if r.currentIndex+1 >= len(r.iterators) {
		r.currentIndex = 0
	} else {
		r.currentIndex++
	}
	return r.currentIndex
}

func (r *concurrentMultiIterator[R]) hasNext() bool {
	for i := range r.iterators {
		if r.iterators[i].hasNext() {
			return true
		}
	}
	return false
}

func (r *concurrentMultiIterator[R]) avaliable() bool {
	return r.hasNext()
}

func (r *concurrentMultiIterator[R]) exeNext() (R, error) {
	for r.hasNext() {
		i := r.nextIndex()

		if !r.iterators[i].hasNext() || !r.iterators[i].avaliable() {
			continue
		}

		return r.iterators[i].exeNext()
	}

	return *new(R), fmt.Errorf("sequence have no next function to execute")
}

func (r *concurrentMultiIterator[R]) exeNextAsAny() (any, error) {
	return r.exeNext()
}
