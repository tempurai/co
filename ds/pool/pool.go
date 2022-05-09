package pool

import (
	"sync"
	"sync/atomic"
)

type poolBasic[K any] struct {
	doneWG   sync.WaitGroup
	jobCache sync.Pool

	seq uint64
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
