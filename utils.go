package co

import (
	"sync"
	"sync/atomic"
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

type AtomicBool struct {
	flag int32
}

func (b *AtomicBool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}

	atomic.StoreInt32(&(b.flag), i)
}

func (b *AtomicBool) Get() bool {
	return atomic.LoadInt32(&(b.flag)) != 0
}
