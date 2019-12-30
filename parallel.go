package co

import (
	"container/list"
	"sync"
)

func NewParallel(workers int) *parallelDispatcher {
	d := &parallelDispatcher{
		workerPool: make(chan chan payload, workers),
		quit:       make(chan bool),

		maxWorkers: workers,

		queueCond: sync.Cond{L: NewLockedMutex()},
		queue:     list.New(),

		wg: &sync.WaitGroup{},
	}

	for i := 0; i < d.maxWorkers; i++ {
		worker := newParallelWorker(d.workerPool, d.wg)
		d.workers = append(d.workers, worker)
		worker.start()
	}

	d.listen()
	return d
}

type payload func() ()

type parallelDispatcher struct {
	workerPool chan chan payload // received empty job queue
	quit       chan bool

	maxWorkers int
	workers    []parallelWorker

	mux       sync.Mutex // mutex of queue
	queueCond sync.Cond // condition variable for queue
	queue     *list.List

	wg *sync.WaitGroup
}

func (d *parallelDispatcher) listen() {
	go func() {
		for {
			select {
			case <-d.quit:
				return

			case jobChannel := <-d.workerPool:
				if d.queue.Len() == 0 { // if no data avaliable, wait
					d.queueCond.Wait()
				}
				if d.queue.Len() == 0 { // which means unlock but still no data
					if quit, ok := ReadBoolChan(d.quit); quit && ok {
						return
					}
					continue
				}

				d.mux.Lock()

				el := d.queue.Front()
				jobChannel <- el.Value.(payload)
				d.queue.Remove(el)

				d.mux.Unlock()
			}
		}
	}()
}

func (d *parallelDispatcher) Add(job payload) {
	d.wg.Add(1)

	d.mux.Lock()
	defer d.mux.Unlock()

	d.queue.PushBack(job)
	d.queueCond.Signal()
}

func (d *parallelDispatcher) Wait() {
	d.wg.Wait()
	d.Stop()
}

func (d *parallelDispatcher) Stop() {
	for _, worker := range d.workers {
		worker.stop()
	}
	go func() {
		d.quit <- true
	}()
	d.queueCond.Broadcast()
}

type parallelWorker struct {
	workerPool chan chan payload
	jobChannel chan payload
	quit       chan bool
	wg         *sync.WaitGroup
}

func newParallelWorker(workerPool chan chan payload, wg *sync.WaitGroup) parallelWorker {
	return parallelWorker{
		workerPool: workerPool,
		jobChannel: make(chan payload),
		quit:       make(chan bool),
		wg:         wg,
	}
}

func (w parallelWorker) start() {
	go func() {
		for {
			w.workerPool <- w.jobChannel

			select {
			case <-w.quit:
				return

			case job := <-w.jobChannel:
				job() // fire
				w.wg.Done()
			}
		}
	}()
}

func (w parallelWorker) stop() {
	w.quit <- true
}
