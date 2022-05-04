package pool

import (
	"sync"
	"sync/atomic"

	"go.tempura.ink/co/ds/queue"
	co_sync "go.tempura.ink/co/internal/sync"
)

func NewDispatchPool[K any](maxWorkers int) *DispatcherPool[K] {
	p := &DispatcherPool[K]{
		poolBasic:      newPoolBasic[K](),
		quitCh:         make(chan bool),
		idleDispatcher: int32(maxWorkers),

		workerCond: co_sync.NewCondCh(&sync.Mutex{}),
		jobQueue:   queue.NewQueue[*job[K]](),
	}

	p.pool.New = func() any {
		return NewDispatcher(&p.doneCh)
	}

	p.startListening()
	return p
}

type DispatcherPool[K any] struct {
	*poolBasic[K]

	pool           sync.Pool
	workerCond     *co_sync.CondCh
	idleDispatcher int32

	callbackFn func(id uint64, val K)
	quitCh     chan bool

	jobQueue *queue.Queue[*job[K]]
}

func (p *DispatcherPool[K]) startListening() {
	co_sync.SafeGo(func() {
		for {
			select {
			case <-p.quitCh:
				return

			case data := <-p.doneCh:
				p.workerCond.Signal(func() {
					p.pool.Put(data.workerRef)
					atomic.AddInt32(&p.idleDispatcher, 1)
				})

				if p.callbackFn != nil {
					co_sync.SafeGo(func() {
						p.callbackFn(data.seq, data.val)
					})
				}
				p.doneWG.Done()
			}
		}
	})

	co_sync.SafeGo(func() {
		for {
			select {
			case <-p.quitCh:
				return
			case <-p.workerCond.WaitCh(func() bool { return p.jobQueue.Len() == 0 || p.idleDispatcher == 0 }):
				atomic.AddInt32(&p.idleDispatcher, -1)

				w := p.pool.Get().(*Dispatcher[K])
				w.trigger(p.jobQueue.Dequeue())
			}
		}
	})
}

func (p *DispatcherPool[K]) SetCallbackFn(fn func(uint64, K)) *DispatcherPool[K] {
	p.callbackFn = fn
	return p
}

func (p *DispatcherPool[K]) AddJobAt(seq uint64, fn func() K) uint64 {
	p.workerCond.Signal(func() {
		p.jobQueue.Enqueue(&job[K]{fn: fn, seq: seq})
	})

	p.doneWG.Add(1)
	return seq
}

func (p *DispatcherPool[K]) AddJob(fn func() K) uint64 {
	id := atomic.AddUint64(&p.seq, 1)
	return p.AddJobAt(id, fn)
}

func (p *DispatcherPool[K]) Wait() *DispatcherPool[K] {
	p.doneWG.Wait()
	return p
}

func (p *DispatcherPool[K]) Stop() {
	p.quitCh <- true
}

type Dispatcher[K any] struct {
	doneCh *chan *jobDone[K]
}

func NewDispatcher[K any](doneCh *chan *jobDone[K]) *Dispatcher[K] {
	w := Dispatcher[K]{
		doneCh: doneCh,
	}
	return &w
}

func (w *Dispatcher[K]) trigger(load *job[K]) {
	co_sync.SafeGo(func() {
		*(w.doneCh) <- &jobDone[K]{val: load.fn(), seq: load.seq, workerRef: w}
	})
}
