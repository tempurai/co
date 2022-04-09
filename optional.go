package co

import (
	"fmt"
	"reflect"
)

type Optional[R any] struct {
	data  R
	valid bool
}

func NewOptionalEmpty[R any]() *Optional[R] {
	return &Optional[R]{
		data:  *new(R),
		valid: false,
	}
}

func OptionalOf[R any](val R) *Optional[R] {
	return &Optional[R]{
		data:  val,
		valid: true,
	}
}

func OptionalIf[R any](val R) *Optional[R] {
	return &Optional[R]{
		data:  val,
		valid: !reflect.ValueOf(val).IsZero(),
	}
}

func (op Optional[R]) String() string {
	if op.valid {
		str := fmt.Sprintf("%v", op.data)
		return str
	} else {
		return "empty optional"
	}
}

func (op Optional[R]) AsOptional() *Optional[any] {
	op2 := OptionalOf[any](op.data)
	op2.valid = op.valid
	return op2
}

func (op Optional[R]) Get() R {
	return op.data
}

func (op Optional[R]) IsPresent() bool {
	return op.valid
}

func (op Optional[R]) Equals(value R) bool {
	if op.valid {
		return reflect.DeepEqual(value, op.data)
	}
	return false
}
