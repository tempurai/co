package co_sync

import "sync"

type CondCh struct {
	*sync.Cond
}

func NewCondCh(mux *sync.Mutex) *CondCh {
	return &CondCh{Cond: sync.NewCond(mux)}
}

func (c *CondCh) Signal(fn func()) {
	CondSignal(c.Cond, fn)
}

func (c *CondCh) Broadcast(fn func()) {
	CondBroadcast(c.Cond, fn)
}

func (c *CondCh) Wait(fn func() bool) {
	CondWait(c.Cond, fn)
}

func (c *CondCh) WaitCh(fn func() bool) <-chan struct{} {
	ch := chanPool.Get().(chan struct{})
	SafeGo(func() {
		c.Cond.L.Lock()
		for fn() {
			c.Cond.Wait()
		}
		c.Cond.L.Unlock()
		ch <- struct{}{}
		chanPool.Put(ch)
	})
	return ch
}
