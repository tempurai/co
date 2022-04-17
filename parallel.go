package co

import (
	"sync"
	"sync/atomic"

	"tempura.ink/co/ds/pool"
	co_sync "tempura.ink/co/internal/sync"
)

type parallel[R any] struct {
	workerPool *pool.WorkerPool[R]
	seqFnMap   sync.Map

	ifPersistentData bool
	storedData       *List[R]

	sent         uint64
	finished     uint64
	recieverCond *sync.Cond
}

func NewParallel[R any](maxWorkers int) *parallel[R] {
	d := &parallel[R]{
		workerPool:   pool.NewWorkerPool[R](maxWorkers),
		storedData:   NewList[R](),
		recieverCond: sync.NewCond(&sync.Mutex{}),
	}

	d.workerPool.SetCallbackFn(d.receiveValue)
	return d
}

func (d *parallel[R]) SetPersistentData(b bool) *parallel[R] {
	d.ifPersistentData = b
	return d
}

func (d *parallel[R]) Process(fn func() R) chan R {
	atomic.AddUint64(&d.sent, 1)
	seq := d.workerPool.ReserveSeq()
	ch := make(chan R)
	d.seqFnMap.Store(seq, ch)

	d.workerPool.AddJobAt(seq, fn)
	return ch
}

func (d *parallel[R]) receiveValue(seq uint64, val R) {
	//TODO: check if seq is in map, if not then use cond
	ch, _ := d.seqFnMap.Load(seq)
	co_sync.SafeGo(func() {
		co_sync.SafeSend(ch.(chan R), val)
	})
	d.seqFnMap.Delete(seq)

	if d.ifPersistentData {
		d.storedData.setAt(int(seq)-1, val)
	}

	co_sync.CondBoardcast(d.recieverCond, func() {
		atomic.AddUint64(&d.finished, 1)
	})
}

func (d *parallel[R]) GetData() []R {
	if !d.ifPersistentData {
		panic("co/paralle error when get data: persistent data mode is not set")
	}
	return d.storedData.list
}

func (d *parallel[R]) Wait() *parallel[R] {
	d.workerPool.Wait()
	co_sync.CondWait(d.recieverCond, func() bool {
		return d.finished != d.sent
	})
	return d
}
