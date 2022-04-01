package co

import (
	"fmt"
)

type AsyncMergedSequence[R any] struct {
	*asyncSequence[R]

	aSequenceables []AsyncSequenceable[R]
}

func NewAsyncMergedSequence[R any](as ...AsyncSequenceable[R]) *AsyncMergedSequence[R] {
	a := &AsyncMergedSequence[R]{
		aSequenceables: as,
	}
	a.asyncSequence = NewAsyncSequence(a.Iterator())
	return a
}

func (c *AsyncMergedSequence[R]) Iterator() Iterator[R] {
	it := &asyncMergedSequenceIterator[R]{}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	for i := range c.aSequenceables {
		it.its = append(it.its, c.aSequenceables[i].Iterator())
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
