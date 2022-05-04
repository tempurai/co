package syncx

import "sync/atomic"

type AtomicBool struct {
	flag int32
}

func (b *AtomicBool) getInt32(value bool) int32 {
	var i int32 = 0
	if value {
		i = 1
	}
	return i
}

func (b *AtomicBool) Set(value bool) {
	i := b.getInt32(value)
	atomic.StoreInt32(&(b.flag), i)
}

func (b *AtomicBool) Get() bool {
	return atomic.LoadInt32(&(b.flag)) != 0
}
