package pool

import (
	"runtime"
	"sync/atomic"

	"go.tempura.ink/co/ds/queue"
	syncx "go.tempura.ink/co/internal/syncx"
)

func NewWorkerPool[K any](maxWorkers int) *WorkerPool[K] {
	p := &WorkerPool[K]{
		poolBasic: newPoolBasic[K](),

		workers:   make([]*Worker[K], maxWorkers),
		loadQueue: queue.NewQueue[*job[K]](),
	}

	for i := 0; i < maxWorkers; i++ {
		worker := NewWorker(p)
		p.workers = append(p.workers, worker)
	}

	return p
}

type WorkerPool[K any] struct {
	*poolBasic[K]

	workers []*Worker[K]

	pendingJob uint32
	loadQueue  *queue.Queue[*job[K]]
}

func (p *WorkerPool[K]) AddJobAt(seq uint64, fn func() K) uint64 {
	p.doneWG.Add(1)

	load := p.jobCache.Get().(*job[K])
	load.fn, load.seq = fn, seq

	p.loadQueue.Enqueue(load)
	atomic.AddUint32(&p.pendingJob, 1)

	return seq
}

func (p *WorkerPool[K]) AddJob(fn func() K) uint64 {
	s := p.ReserveSeq()
	return p.AddJobAt(s, fn)
}

func (p *WorkerPool[K]) Stop() {
	for _, worker := range p.workers {
		worker.quit = true
	}
}

type Worker[K any] struct {
	pool *WorkerPool[K]
	quit bool
}

func NewWorker[K any](p *WorkerPool[K]) *Worker[K] {
	w := &Worker[K]{
		pool: p,
	}
	w.startListening()
	return w
}

func (w *Worker[K]) startListening() {
	syncx.SafeGo(func() {
		for {
			yeildCost := 1
			for !w.quit {
				currentPendingJob := w.pool.pendingJob
				if currentPendingJob > 0 {
					if atomic.CompareAndSwapUint32(&w.pool.pendingJob, currentPendingJob, currentPendingJob-1) {
						break
					}
				}
				for i := 0; i < yeildCost; i++ {
					runtime.Gosched()
				}
				yeildCost <<= 1
			}

			if w.quit {
				return
			}

			var load *job[K] = w.pool.loadQueue.Dequeue()
			val := load.fn()

			w.pool.callCallback(load.seq, val)

			w.pool.jobCache.Put(load)
			w.pool.doneWG.Done()
		}
	})
}
