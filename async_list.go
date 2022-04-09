package co

import (
	"sync"
	"sync/atomic"
)

type List[R any] struct {
	list []R

	rwmux sync.RWMutex
}

func NewList[R any]() *List[R] {
	return &List[R]{list: make([]R, 0)}
}

func (l *List[R]) len() int {
	l.rwmux.RLock()
	defer l.rwmux.RUnlock()

	return len(l.list)
}

func (l *List[R]) popFirst() R {
	l.rwmux.Lock()
	defer l.rwmux.Unlock()

	var result R
	result, l.list = l.list[0], l.list[1:]
	return result
}

func (l *List[R]) getAt(i int) R {
	l.rwmux.RLock()
	defer l.rwmux.RUnlock()

	return l.list[i]
}

func (l *List[R]) setAt(i int, val R) {
	l.rwmux.Lock()
	defer l.rwmux.Unlock()

	l.resizeToInternal(i + 1)
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

func (l *List[R]) resizeTo(size int) {
	l.rwmux.Lock()
	defer l.rwmux.Unlock()

	l.resizeToInternal(size)
}

func (l *List[R]) resizeToInternal(size int) {
	if size <= len(l.list) {
		return
	}
	l.list = append(l.list, make([]R, size-len(l.list))...)
}

type iterativeList[R any] struct {
	List[R]
}

func NewIterativeList[R any]() *iterativeList[R] {
	return &iterativeList[R]{
		List: List[R]{list: make([]R, 0)},
	}
}

func (l *iterativeList[R]) iterator() *iterativeListIterator[R] {
	it := &iterativeListIterator[R]{
		iterativeList: l,
		currentIndex:  0,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type iterativeListIterator[R any] struct {
	*asyncSequenceIterator[R]

	*iterativeList[R]
	currentIndex int32
}

func (it *iterativeListIterator[R]) preflight() bool {
	return atomic.LoadInt32(&it.currentIndex) < int32(it.len())
}

func (it *iterativeListIterator[R]) next() (*Optional[R], error) {
	if !it.preflight() {
		return NewOptionalEmpty[R](), nil
	}

	defer func() { atomic.AddInt32(&it.currentIndex, 1) }()
	return OptionalOf(it.list[it.currentIndex]), nil
}
