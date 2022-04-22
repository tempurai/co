package pool

import (
	"sync"
	"sync/atomic"

	"go.tempura.ink/co/ds/queue"
	co_sync "go.tempura.ink/co/internal/sync"
)

func NewWorkerPool[K any](maxWorkers int) *WorkerPool[K] {
	p := &WorkerPool[K]{
		poolBasic: newPoolBasic[K](),

		workers:     make([]*Worker[K], maxWorkers),
		quitCh:      make(chan bool),
		idleWorkers: queue.NewQueue[*Worker[K]](),

		workerCond: sync.NewCond(&sync.Mutex{}),
		jobQueue:   queue.NewQueue[*job[K]](),
	}

	for i := 0; i < maxWorkers; i++ {
		worker := NewWorker(uint16(i), &p.doneCh)
		p.idleWorkers.Enqueue(worker)
		p.workers[i] = worker
	}

	p.startListening()
	return p
}

type WorkerPool[K any] struct {
	*poolBasic[K]

	workerCond  *sync.Cond
	idleWorkers *queue.Queue[*Worker[K]]

	callbackFn func(id uint64, val K)

	workers []*Worker[K] // for shutting down
	quit    bool
	quitCh  chan bool

	jobQueue *queue.Queue[*job[K]]
}

func (p *WorkerPool[K]) startListening() {
	co_sync.SafeGo(func() {
		for {
			select {
			case <-p.quitCh:
				return

			case done := <-p.doneCh:
				co_sync.CondSignal(p.workerCond, func() {
					p.idleWorkers.Enqueue(done.workerRef.(*Worker[K]))
				})

				if p.callbackFn != nil {
					co_sync.SafeGo(func() {
						p.callbackFn(done.seq, done.val)
					})
				}
				p.doneWG.Done()
			}
		}
	})

	co_sync.SafeGo(func() {
		for {
			co_sync.CondWait(p.workerCond, func() bool {
				return !p.quit && (p.jobQueue.Len() == 0 || p.idleWorkers.Len() == 0)
			})
			if p.quit {
				return
			}

			w := p.idleWorkers.Dequeue()
			w.jobCh <- p.jobQueue.Dequeue()
		}
	})
}

func (p *WorkerPool[K]) SetCallbackFn(fn func(uint64, K)) *WorkerPool[K] {
	p.callbackFn = fn
	return p
}

func (p *WorkerPool[K]) ReserveSeq() uint64 {
	return atomic.AddUint64(&p.seq, 1)
}

func (p *WorkerPool[K]) AddJobAt(seq uint64, fn func() K) uint64 {
	co_sync.CondSignal(p.workerCond, func() {
		p.jobQueue.Enqueue(&job[K]{fn: fn, seq: seq})
	})

	p.doneWG.Add(1)
	return seq
}

func (p *WorkerPool[K]) AddJob(fn func() K) uint64 {
	id := atomic.AddUint64(&p.seq, 1)

	co_sync.CondSignal(p.workerCond, func() {
		p.jobQueue.Enqueue(&job[K]{fn: fn, seq: id})
	})

	p.doneWG.Add(1)
	return id
}

func (p *WorkerPool[K]) Wait() {
	p.doneWG.Wait()
}

func (p *WorkerPool[K]) Stop() {
	for _, worker := range p.workers {
		worker.stop()
	}
	p.quitCh <- true
	p.quit = true
	p.workerCond.Broadcast()
}

type Worker[K any] struct {
	id     uint16
	jobCh  chan *job[K]
	quitCh chan bool
	doneCh *chan *jobDone[K]
}

func NewWorker[K any](ID uint16, doneCh *chan *jobDone[K]) *Worker[K] {
	w := &Worker[K]{
		id:     ID,
		jobCh:  make(chan *job[K]),
		quitCh: make(chan bool),
		doneCh: doneCh,
	}
	w.startListening()
	return w
}

func (w *Worker[K]) startListening() {
	co_sync.SafeGo(func() {
		for {
			select {
			case <-w.quitCh:
				return

			case load := <-w.jobCh:
				*(w.doneCh) <- &jobDone[K]{val: load.fn(), seq: load.seq, workerRef: w}
			}
		}
	})
}

func (w *Worker[K]) stop() {
	w.quitCh <- true
}
