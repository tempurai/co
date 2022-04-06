package co_sync

import (
	"fmt"
	"runtime/debug"
	"sync"
)

func SafeSend[T any](ch chan T, value T) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
			fmt.Printf("channel %+v send out %+v failed\n", ch, value)
		}
	}()

	ch <- value
	return false
}

func SafeNSend[T any](ch chan T, value T) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
			fmt.Printf("channel %+v send out %+v failed\n", ch, value)
		}
	}()

	select {
	case ch <- value:
	default:
		fmt.Printf("channel %+v send out %+v blocked omitted, stacktrace: %+v \n", ch, value, string(debug.Stack()))
	}
	return false
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

func CondSignal(cond *sync.Cond, fn func()) {
	cond.L.Lock()
	fn()
	cond.Signal()
	cond.L.Unlock()
}

func CondBoardcast(cond *sync.Cond, fn func()) {
	cond.L.Lock()
	fn()
	cond.Broadcast()
	cond.L.Unlock()
}

func CondWait(cond *sync.Cond, fn func() bool) {
	cond.L.Lock()
	for fn() {
		cond.Wait()
	}
	cond.L.Unlock()
}
