package syncx

import "sync"

type Condx struct {
	*sync.Cond
}

func NewCondx(mux *sync.Mutex) *Condx {
	return &Condx{Cond: sync.NewCond(mux)}
}

func (c *Condx) Signalify(fn func()) {
	CondSignal(c.Cond, fn)
}

func (c *Condx) Broadcastify(fn func()) {
	CondBroadcast(c.Cond, fn)
}

func (c *Condx) Waitify(fn1 func() bool, fn2 func()) {
	CondWaitWrap(c.Cond, fn1, fn2)
}

func (c *Condx) WaitCh(fn func() bool) <-chan struct{} {
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
