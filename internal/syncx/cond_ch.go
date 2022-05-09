package syncx

import "sync"

type CondCh struct {
	*sync.Cond
}

func NewCondCh(mux *sync.Mutex) *CondCh {
	return &CondCh{Cond: sync.NewCond(mux)}
}

func (c *CondCh) Signalify(fn func()) {
	CondSignal(c.Cond, fn)
}

func (c *CondCh) Broadcastify(fn func()) {
	CondBroadcast(c.Cond, fn)
}

func (c *CondCh) Waitify(fn1 func() bool, fn2 func()) {
	CondWaitWrap(c.Cond, fn1, fn2)
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
