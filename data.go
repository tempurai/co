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

func NewDataWith[R any](val R, err error) *data[R] {
	return &data[R]{val, err}
}

type determinedDataList[R any] struct {
	*List[*data[R]]

	rwmux sync.RWMutex
}
