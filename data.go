package co

import (
	"sync"
)

type data[R any] struct {
	value R
	err   error
}

func (d *data[R]) GetValue() R {
	return d.value
}

func (d *data[R]) GetError() error {
	return d.err
}

func NewData[R any]() *data[R] {
	return &data[R]{}
}

type determinedDataList[R any] struct {
	*List[*data[R]]

	rwmux sync.RWMutex
}

func NewDeterminedDataList[R any]() *determinedDataList[R] {
	return &determinedDataList[R]{
		List: NewList[*data[R]](),
	}
}

func (d *determinedDataList[R]) add(val R, err error) {
	d.rwmux.RLock()
	defer d.rwmux.RUnlock()

	d.List.add(&data[R]{val, err})
}

func (d *determinedDataList[R]) getAt(i int) (R, error) {
	d.rwmux.RLock()
	defer d.rwmux.RUnlock()

	return d.List.list[i].value, d.List.list[i].err
}

func (d *determinedDataList[R]) getDataAt(i int) *data[R] {
	d.rwmux.RLock()
	defer d.rwmux.RUnlock()

	return d.List.list[i]
}

func (d *determinedDataList[R]) setAt(i int, val R, err error) {
	d.rwmux.Lock()
	defer d.rwmux.Unlock()

	d.resizeTo(i + 1)
	d.List.list[i] = &data[R]{val, err}
}

func (d *determinedDataList[R]) append(list *determinedDataList[R]) {
	d.rwmux.RLock()
	defer d.rwmux.RUnlock()

	d.List.add(list.list...)
}
