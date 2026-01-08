package co

import (
	"sync"
	"sync/atomic"

	"github.com/tempurai/co/ds/pool"
	syncx "github.com/tempurai/co/internal/syncx"
)

type parallel[R any] struct {
	workerPool *pool.WorkerPool[R]
	seqFnMap   sync.Map

	ifPersistentData bool
	storedData       *List[R]

	sent         uint64
	finished     uint64
	recieverCond *syncx.Condx
}

func NewParallel[R any](maxWorkers int) *parallel[R] {
	d := &parallel[R]{
		workerPool:   pool.NewWorkerPool[R](maxWorkers),
		storedData:   NewList[R](),
		recieverCond: syncx.NewCondx(&sync.Mutex{}),
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
	ch, ok := d.seqFnMap.Load(seq)
	if ok {
		syncx.SafeNSend(ch.(chan R), val)
	}

	d.seqFnMap.Delete(seq)

	if d.ifPersistentData {
		d.storedData.setAt(int(seq)-1, val)
	}

	d.recieverCond.Broadcastify(&syncx.BroadcastOption{
		PreProcessFn: func() {
			atomic.AddUint64(&d.finished, 1)
		}},
	)
}

func (d *parallel[R]) GetData() []R {
	if !d.ifPersistentData {
		panic("co/paralle error when get data: persistent data mode is not set")
	}
	return d.storedData.list
}

func (d *parallel[R]) Wait() *parallel[R] {
	d.workerPool.Wait()
	d.recieverCond.Waitify(&syncx.WaitOption{
		ConditionFn: func() bool {
			return d.finished != d.sent
		},
	})

	return d
}
