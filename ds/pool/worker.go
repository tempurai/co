package pool

import (
	"sync"
	"sync/atomic"

	"go.tempura.ink/co/ds/queue"
	syncx "go.tempura.ink/co/internal/syncx"
)

func NewWorkerPool[K any](maxWorkers int) *WorkerPool[K] {
	p := &WorkerPool[K]{
		poolBasic: newPoolBasic[K](),

		workers: make([]*Worker[K], maxWorkers),

		loadCond:  syncx.NewCondx(&sync.Mutex{}),
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

	callbackFn func(id uint64, val K)

	workers []*Worker[K] // for shutting down
	quit    bool

	loadCond  *syncx.Condx
	loadQueue *queue.Queue[*job[K]]
}

func (p *WorkerPool[K]) SetCallbackFn(fn func(uint64, K)) *WorkerPool[K] {
	p.callbackFn = fn
	return p
}

func (p *WorkerPool[K]) AddJobAt(seq uint64, fn func() K) uint64 {
	p.doneWG.Add(1)

	p.loadCond.Signalify(func() {
		load := p.jobCache.Get().(*job[K])
		load.fn, load.seq = fn, seq

		p.loadQueue.Enqueue(load)
	})

	return seq
}

func (p *WorkerPool[K]) AddJob(fn func() K) uint64 {
	id := atomic.AddUint64(&p.seq, 1)

	return p.AddJobAt(id, fn)
}

func (p *WorkerPool[K]) Stop() {
	p.loadCond.Broadcastify(func() {
		for _, worker := range p.workers {
			worker.quit = true
		}
	})
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
			var load *job[K]

			w.pool.loadCond.Waitify(func() bool {
				return !w.quit && w.pool.loadQueue.IsEmpty()
			}, func() {
				load = w.pool.loadQueue.Dequeue()
			})

			if w.quit {
				return
			}
			val := load.fn()

			if w.pool.callbackFn != nil {
				syncx.SafeGo(func() {
					w.pool.callbackFn(load.seq, val)
				})
			}

			w.pool.jobCache.Put(load)
			w.pool.doneWG.Done()
		}
	})
}
