package syncx

import "sync"

type Condx struct {
	*sync.Cond
}

func NewCondx(mux *sync.Mutex) *Condx {
	return &Condx{Cond: sync.NewCond(mux)}
}

type SignalOption struct {
	PreProcessFn  func()
	PostProcessFn func()
}

func (c *Condx) Signalify(op *SignalOption) {
	c.Cond.L.Lock()
	if op.PreProcessFn != nil {
		op.PreProcessFn()
	}
	c.Cond.Signal()
	if op.PostProcessFn != nil {
		op.PostProcessFn()
	}
	c.Cond.L.Unlock()
}

type BroadcastOption struct {
	PreProcessFn  func()
	PostProcessFn func()
}

func (c *Condx) Broadcastify(op *BroadcastOption) {
	c.Cond.L.Lock()
	if op.PreProcessFn != nil {
		op.PreProcessFn()
	}
	c.Cond.Broadcast()
	if op.PostProcessFn != nil {
		op.PostProcessFn()
	}
	c.Cond.L.Unlock()
}

type WaitOption struct {
	ConditionFn   func() bool
	PostProcessFn func()
}

func (c *Condx) Waitify(op *WaitOption) {
	c.Cond.L.Lock()
	for op.ConditionFn() {
		c.Cond.Wait()
	}
	if op.PostProcessFn != nil {
		op.PostProcessFn()
	}
	c.Cond.L.Unlock()
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
