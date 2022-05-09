package pool

import (
	"sync"

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

	workers []*Worker[K] // for shutting down
	quit    bool

	loadCond  *syncx.Condx
	loadQueue *queue.Queue[*job[K]]
}

func (p *WorkerPool[K]) AddJobAt(seq uint64, fn func() K) uint64 {
	p.doneWG.Add(1)

	load := p.jobCache.Get().(*job[K])
	load.fn, load.seq = fn, seq

	p.loadCond.Signalify(&syncx.SignalOption{
		PreProcessFn: func() {
			p.loadQueue.Enqueue(load)
		},
	})

	return seq
}

func (p *WorkerPool[K]) AddJob(fn func() K) uint64 {
	return p.AddJobAt(p.ReserveSeq(), fn)
}

func (p *WorkerPool[K]) Stop() {
	p.loadCond.Broadcastify(&syncx.BroadcastOption{
		PreProcessFn: func() {
			for _, worker := range p.workers {
				worker.quit = true
			}
		}},
	)
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

			w.pool.loadCond.Waitify(&syncx.WaitOption{
				ConditionFn: func() bool {
					return !w.quit && w.pool.loadQueue.IsEmpty()
				},
				PostProcessFn: func() {
					load = w.pool.loadQueue.Dequeue()
				},
			})

			if w.quit {
				return
			}

			val := load.fn()
			w.pool.callCallback(load.seq, val)

			w.pool.jobCache.Put(load)
			w.pool.doneWG.Done()
		}
	})
}
