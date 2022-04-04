package co

import (
	"sync"
	"sync/atomic"
)

func NewHeavyWorkerPool[K any](maxWorkers int) *HeavyWorkerPool[K] {
	p := &HeavyWorkerPool[K]{
		workers:     make([]*Worker[K], maxWorkers),
		quitCh:      make(chan bool),
		doneCh:      make(chan *jobDone[K]),
		idleWorkers: NewQueue[*Worker[K]](),

		workerCond: sync.NewCond(&sync.Mutex{}),
		jobQueue:   NewQueue[*job[K]](),
	}

	for i := 0; i < maxWorkers; i++ {
		worker := NewWorker(uint16(i), &p.doneCh)
		p.idleWorkers.Enqueue(worker)
		p.workers[i] = worker
	}

	p.startListening()
	return p
}

type HeavyWorkerPool[K any] struct {
	workers     []*Worker[K] // for shutting down
	workerCond  *sync.Cond
	idleWorkers *Queue[*Worker[K]]

	doneCh chan *jobDone[K]
	doneWG sync.WaitGroup

	callbackFn func(id uint64, val K)

	quit   bool
	quitCh chan bool

	jobQueue *Queue[*job[K]]

	seq uint64
}

func (p *HeavyWorkerPool[K]) startListening() {
	SafeGo(func() {
		for {
			select {
			case <-p.quitCh:
				return

			case done := <-p.doneCh:
				CondSignal(p.workerCond, func() {
					p.idleWorkers.Enqueue(done.workerRef)
				})

				if p.callbackFn != nil {
					SafeGo(func() {
						p.callbackFn(done.seq, done.val)
					})
				}
				p.doneWG.Done()
			}
		}
	})

	SafeGo(func() {
		for {
			CondWait(p.workerCond, func() bool {
				return p.quit || p.jobQueue.Len() == 0 || p.idleWorkers.Len() == 0
			})
			if p.quit {
				return
			}

			w := p.idleWorkers.Dequeue()
			w.jobCh <- p.jobQueue.Dequeue()
		}
	})
}

func (p *HeavyWorkerPool[K]) AddJob(fn func() K) uint64 {
	id := atomic.AddUint64(&p.seq, 1)

	CondSignal(p.workerCond, func() {
		p.jobQueue.Enqueue(&job[K]{fn: fn, seq: id})
	})

	p.doneWG.Add(1)
	return id
}

func (p *HeavyWorkerPool[K]) Wait() {
	p.doneWG.Wait()
}

func (p *HeavyWorkerPool[K]) Stop() {
	for _, worker := range p.workers {
		worker.stop()
	}
	p.quitCh <- true
	p.quit = true
	p.workerCond.Broadcast()
}

type job[K any] struct {
	fn  func() K
	seq uint64
}

type jobDone[K any] struct {
	val       K
	seq       uint64
	workerRef *Worker[K]
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
	SafeGo(func() {
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
