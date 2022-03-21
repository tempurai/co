package co

import (
	"sync"
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
