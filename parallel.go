package co

import (
	"container/list"
	"sync"
)

func NewParallel(workers int) *parallelDispatcher[int] {
	d := createBaseParallel[int](workers)
	return d
}

func NewParallelWithResponse[K any](workers int) *parallelDispatcher[K] {
	d := createBaseParallel[K](workers)
	d.ifResponse = true
	return d
}

func createBaseParallel[K any](workers int) *parallelDispatcher[K] {
	d := &parallelDispatcher[K]{
		workerPool: make(chan chan payload[K], workers),
		quit:       make(chan bool),

		maxWorkers: workers,

		queueCond: sync.Cond{L: NewLockedMutex()},
		queue:     list.New(),

		wg: &sync.WaitGroup{},
	}

	for i := 0; i < d.maxWorkers; i++ {
		worker := newParallelWorker(d)
		d.workers = append(d.workers, worker)
		worker.start()
	}

	d.listen()
	return d
}

type payload[K any] struct {
	job func() K
	seq int
}

type parallelDispatcher[K any] struct {
	workerPool chan chan payload[K] // received empty job queue
	quit       chan bool

	maxWorkers int
	workers    []parallelWorker[K]

	mux       sync.Mutex // mutex of queue
	queueCond sync.Cond  // condition variable for queue
	queue     *list.List

	wg *sync.WaitGroup

	ifResponse   bool
	responses    []K        // record response
	responsesMux sync.Mutex // prevent concurrent write since no idea size of job
}

func (d *parallelDispatcher[K]) listen() {
	go func() {
		for {
			select {
			case <-d.quit:
				return

			case jobChannel := <-d.workerPool:
				if d.queue.Len() == 0 { // if no data available, wait
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
				jobChannel <- el.Value.(payload[K])
				d.queue.Remove(el)

				d.mux.Unlock()
			}
		}
	}()
}

func (d *parallelDispatcher[K]) Add(job func()) {
	d.AddWithResponse(func() K {
		job()
		return *new(K)
	})
}

func (d *parallelDispatcher[K]) AddWithResponse(job func() K) {
	d.wg.Add(1)

	if d.ifResponse {
		d.responsesMux.Lock()
		d.responses = append(d.responses, *new(K))
		d.responsesMux.Unlock()
	}

	d.mux.Lock()
	d.queue.PushBack(payload[K]{job: job, seq: len(d.responses) - 1})
	d.mux.Unlock()

	d.queueCond.Signal()
}

func (d *parallelDispatcher[K]) Wait() []K {
	d.wg.Wait()
	d.Stop()
	return d.responses
}

func (d *parallelDispatcher[K]) Stop() {
	for _, worker := range d.workers {
		worker.stop()
	}
	go func() {
		d.quit <- true
	}()
	d.queueCond.Broadcast()
}

type parallelWorker[K any] struct {
	dispatcher *parallelDispatcher[K]
	jobChannel chan payload[K]
	quit       chan bool
}

func newParallelWorker[K any](d *parallelDispatcher[K]) parallelWorker[K] {
	return parallelWorker[K]{
		dispatcher: d,
		jobChannel: make(chan payload[K]),
		quit:       make(chan bool),
	}
}

func (w parallelWorker[K]) start() {
	go func() {
		for {
			w.dispatcher.workerPool <- w.jobChannel

			select {
			case <-w.quit:
				return

			case load := <-w.jobChannel:
				resp := load.job() // fire

				if w.dispatcher.ifResponse {
					w.dispatcher.responsesMux.Lock()
					w.dispatcher.responses[load.seq] = resp
					w.dispatcher.responsesMux.Unlock()
				}

				w.dispatcher.wg.Done()
			}
		}
	}()
}

func (w parallelWorker[K]) stop() {
	w.quit <- true
}
