package co

import (
	"fmt"
)

type AsyncMergedSequence[R any] struct {
	concurrents []AsyncExecutable[R]
}

func NewAsyncMergedSequence[R any](cos ...AsyncExecutable[R]) *AsyncMergedSequence[R] {
	return &AsyncMergedSequence[R]{
		concurrents: cos,
	}
}

func (c *AsyncMergedSequence[R]) Iterator() *asyncMergedSequenceIterator[R] {
	it := &asyncMergedSequenceIterator[R]{}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	for i := range c.concurrents {
		it.its = append(it.its, c.concurrents[i].defaultIterator())
	}
	return it
}

type asyncMergedSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	its          []Iterator[R]
	currentIndex int
}

func (it *asyncMergedSequenceIterator[R]) nextIndex() int {
	if it.currentIndex+1 >= len(it.its) {
		it.currentIndex = 0
	} else {
		it.currentIndex++
	}
	return it.currentIndex
}

func (it *asyncMergedSequenceIterator[R]) preflight() bool {
	for i := range it.its {
		if it.its[i].preflight() {
			return true
		}
	}
	return false
}

func (it *asyncMergedSequenceIterator[R]) consume() (R, error) {
	for it.preflight() {
		idx := it.nextIndex()

		if !it.its[idx].preflight() {
			continue
		}
		return it.its[idx].consume()
	}
	return *new(R), fmt.Errorf("sequence have no consume function to execute")
}
