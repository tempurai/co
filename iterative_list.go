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

func (d *List[R]) len() int {
	return len(d.list)
}

func (d *List[R]) getAt(i int) R {
	d.rwmux.RLock()
	defer d.rwmux.RUnlock()

	return d.list[i]
}

func (d *List[R]) setAt(i int, val R) {
	d.rwmux.Lock()
	defer d.rwmux.Unlock()

	d.list[i] = val
}

func (d *List[R]) add(items ...R) {
	d.rwmux.Lock()
	defer d.rwmux.Unlock()

	d.list = append(d.list, items...)
}

func (d *List[R]) swap(items []R) {
	d.rwmux.Lock()
	defer d.rwmux.Unlock()

	d.list = items
}

func (d *List[R]) resizeTo(l int) {
	if l <= d.len() {
		return
	}

	d.rwmux.Lock()
	defer d.rwmux.Unlock()

	d.list = append(d.list, make([]R, l-d.len())...)
}

type iterativeList[R any] struct {
	List[R]
}

func NewIterativeList[R any]() *iterativeList[R] {
	return &iterativeList[R]{
		List: List[R]{list: make([]R, 0)},
	}
}

func (d *iterativeList[R]) Iterator() *iterativeListIterator[R] {
	return &iterativeListIterator[R]{iterativeList: d}
}

type iterativeListIterator[R any] struct {
	*iterativeList[R]
	currentIndex int
}

func (d *iterativeListIterator[R]) avaliable() bool {
	return d.hasNext()
}

func (d *iterativeListIterator[R]) hasNext() bool {
	return d.currentIndex < d.len()
}

func (d *iterativeListIterator[R]) next() R {
	d.currentIndex++
	return d.list[d.currentIndex-1]
}
