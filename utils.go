package co

import (
	"sync"

	"golang.org/x/exp/constraints"
)

func NewLockedMutex() *sync.Mutex {
	mux := &sync.Mutex{}
	mux.Lock()
	return mux
}

func ReadBoolChan(ch chan bool) (bool, bool) {
	select {
	case x, ok := <-ch:
		if ok {
			return x, ok
		} else {
			return false, false
		}
	default:
		return false, true
	}
}

func SafeSend[T any](ch chan T, value T) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	ch <- value
	return false
}

func CastOrNil[T any](el any) T {
	if (el == nil) {
		return *new(T)
	}
	return el.(T
}

func EvertGET[T constraints](ele []T, target T) bool {
	for _, e := range ele {
		if e <= target {
			return false
		}
	}
	return true
}
