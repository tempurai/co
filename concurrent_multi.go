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

func (c *ConcurrentMulti[R]) Iterator() *concurrentMultiIterator[R] {
	it := &concurrentMultiIterator[R]{}
	for i := range c.concurrents {
		it.iterators = append(it.iterators, c.concurrents[i].defaultIterator())
	}
	return it
}

type concurrentMultiIterator[R any] struct {
	iterators []ExecutableIterator[R]

	currentIndex int
}

func (it *concurrentMultiIterator[R]) nextIndex() int {
	if it.currentIndex+1 >= len(it.iterators) {
		it.currentIndex = 0
	} else {
		it.currentIndex++
	}
	return it.currentIndex
}

func (it *concurrentMultiIterator[R]) hasNext() bool {
	for i := range it.iterators {
		if it.iterators[i].hasNext() {
			return true
		}
	}
	return false
}

func (it *concurrentMultiIterator[R]) available() bool {
	return it.hasNext()
}

func (it *concurrentMultiIterator[R]) exeNext() (R, error) {
	for it.hasNext() {
		i := it.nextIndex()

		if !it.iterators[i].hasNext() || !it.iterators[i].available() {
			continue
		}

		return it.iterators[i].exeNext()
	}

	return *new(R), fmt.Errorf("sequence have no next function to execute")
}

func (it *concurrentMultiIterator[R]) exeNextAsAny() (any, error) {
	return it.exeNext()
}
