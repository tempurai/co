package syncx

import (
	"fmt"
	"runtime/debug"
)

var (
	Debug = false
)

func SafeSend[T any](ch chan T, value T) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
			if Debug {
				fmt.Printf("channel %+v send out %+v failed\n", ch, value)
			}
		}
	}()

	ch <- value
	return false
}

func SafeNSend[T any](ch chan T, value T) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
			if Debug {
				fmt.Printf("channel %+v send out %+v failed\n", ch, value)
			}
		}
	}()

	select {
	case ch <- value:
	default:
		if Debug {
			fmt.Printf("channel %+v send out %+v blocked omitted, stacktrace: %+v \n", ch, value, string(debug.Stack()))
		}
	}
	return false
}

func SafeNRead[T any](ch chan T) T {
	select {
	case val := <-ch:
		return val
	default:
		return *new(T)
	}
}

func SafeClose[T any](ch chan T) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = false
		}
	}()

	close(ch)
	return true
}

func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("go func panic: %+v, stacktrace: %+v \n", r, string(debug.Stack()))
			}
		}()
		fn()
	}()
}

func SafeFn[E any](fn func() E) (val E, err error) {
	if r := recover(); r != nil {
		err = fmt.Errorf("go func panic: %+v, stacktrace: %+v", r, string(debug.Stack()))
	}
	val = fn()
	return
}

func SafeEFn[E any](fn func() (E, error)) (val E, err error) {
	if r := recover(); r != nil {
		err = fmt.Errorf("%w \n go func panic: %+v, stacktrace: %+v", err, r, string(debug.Stack()))
	}
	val, err = fn()
	return
}
