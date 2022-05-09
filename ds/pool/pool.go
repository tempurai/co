package pool

import (
	"sync"
	"sync/atomic"

	syncx "go.tempura.ink/co/internal/syncx"
)

type poolBasic[K any] struct {
	doneWG   sync.WaitGroup
	jobCache sync.Pool

	callbackFn func(id uint64, val K)
	seq        uint64
}

func newPoolBasic[K any]() *poolBasic[K] {
	p := &poolBasic[K]{}
	p.jobCache.New = func() any {
		return &job[K]{}
	}
	return p
}

func (p *poolBasic[K]) ReserveSeq() uint64 {
	return atomic.AddUint64(&p.seq, 1)
}

func (p *poolBasic[K]) Wait() *poolBasic[K] {
	p.doneWG.Wait()
	return p
}

func (p *poolBasic[K]) callCallback(seq uint64, val K) {
	if p.callbackFn != nil {
		syncx.SafeGo(func() {
			p.callbackFn(seq, val)
		})
	}
}

func (p *WorkerPool[K]) SetCallbackFn(fn func(uint64, K)) *WorkerPool[K] {
	p.callbackFn = fn
	return p
}
