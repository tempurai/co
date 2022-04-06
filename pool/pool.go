package pool

import (
	"sync"
	"sync/atomic"
)

type poolBasic[K any] struct {
	doneCh chan *jobDone[K]
	doneWG sync.WaitGroup

	seq uint64
}

func newPoolBasic[K any]() *poolBasic[K] {
	return &poolBasic[K]{
		doneCh: make(chan *jobDone[K]),
	}
}

func (p *poolBasic[K]) ReserveSeq() uint64 {
	return atomic.AddUint64(&p.seq, 1)
}

func (p *poolBasic[K]) Wait() *poolBasic[K] {
	p.doneWG.Wait()
	return p
}
