package co

import (
	"sync"
)

type List[R any] struct {
	list []R

	rwmux sync.RWMutex
}

func NewList[R any]() *List[R] {
	return &List[R]{list: make([]R, 0)}
}

func (l *List[R]) len() int {
	return len(l.list)
}

func (l *List[R]) getAt(i int) R {
	l.rwmux.RLock()
	defer l.rwmux.RUnlock()

	return l.list[i]
}

func (l *List[R]) setAt(i int, val R) {
	l.rwmux.Lock()
	defer l.rwmux.Unlock()

	l.list[i] = val
}

func (l *List[R]) add(items ...R) {
	l.rwmux.Lock()
	defer l.rwmux.Unlock()

	l.list = append(l.list, items...)
}

func (l *List[R]) swap(items []R) {
	l.rwmux.Lock()
	defer l.rwmux.Unlock()

	l.list = items
}

func (l *List[R]) resizeTo(len int) {
	if len <= l.len() {
		return
	}

	l.rwmux.Lock()
	defer l.rwmux.Unlock()

	l.list = append(l.list, make([]R, len-l.len())...)
}

type iterativeList[R any] struct {
	List[R]
}

func NewIterativeList[R any]() *iterativeList[R] {
	return &iterativeList[R]{
		List: List[R]{list: make([]R, 0)},
	}
}

func (it *iterativeList[R]) Iterator() *iterativeListIterator[R] {
	return &iterativeListIterator[R]{iterativeList: it}
}

type iterativeListIterator[R any] struct {
	*iterativeList[R]
	currentIndex int
}

func (it *iterativeListIterator[R]) available() bool {
	return it.hasNext()
}

func (it *iterativeListIterator[R]) hasNext() bool {
	return it.currentIndex < it.len()
}

func (it *iterativeListIterator[R]) next() R {
	it.currentIndex++
	return it.list[it.currentIndex-1]
}
