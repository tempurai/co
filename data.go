package co

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

func (d *data[R]) set(val R, err error) *data[R] {
	d.value = val
	d.err = err
	return d
}

func NewData[R any]() *data[R] {
	return &data[R]{}
}

func NewDataWith[R any](val R, err error) *data[R] {
	return &data[R]{val, err}
}
