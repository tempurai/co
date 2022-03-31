package co

import (
	"fmt"
)

type CoMultiSequence[R any] struct {
	concurrents []CoExecutable[R]
}

func NewCoMultiSequence[R any](cos ...CoExecutable[R]) *CoMultiSequence[R] {
	return &CoMultiSequence[R]{
		concurrents: cos,
	}
}

func (c *CoMultiSequence[R]) Iterator() *coMultiSequenceIterator[R] {
	it := &coMultiSequenceIterator[R]{}
	for i := range c.concurrents {
		it.its = append(it.its, c.concurrents[i].defaultIterator())
	}
	return it
}

type coMultiSequenceIterator[R any] struct {
	its []Iterator[R]

	currentIndex int
}

func (it *coMultiSequenceIterator[R]) nextIndex() int {
	if it.currentIndex+1 >= len(it.its) {
		it.currentIndex = 0
	} else {
		it.currentIndex++
	}
	return it.currentIndex
}

func (it *coMultiSequenceIterator[R]) hasNext() bool {
	for i := range it.its {
		if it.its[i].hasNext() {
			return true
		}
	}
	return false
}

func (it *coMultiSequenceIterator[R]) next() (R, error) {
	for it.hasNext() {
		idx := it.nextIndex()

		if !it.its[idx].hasNext() {
			continue
		}
		return it.its[idx].next()
	}
	return *new(R), fmt.Errorf("sequence have no next function to execute")
}

func (it *coMultiSequenceIterator[R]) nextAny() (any, error) {
	return it.next()
}
