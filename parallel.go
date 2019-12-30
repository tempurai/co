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
		worker := newParallelWorker(d)
		d.workers = append(d.workers, worker)
		worker.start()
	}

	d.listen()
	return d
}

func NewParallelWithResponse(workers int) *parallelDispatcher {
	d := NewParallel(workers)
	d.ifResponse = true
	return d
}

type payload struct {
	job func() interface{}
	seq int
}

type parallelDispatcher struct {
	workerPool chan chan payload // received empty job queue
	quit       chan bool

	maxWorkers int
	workers    []parallelWorker

	mux       sync.Mutex // mutex of queue
	queueCond sync.Cond  // condition variable for queue
	queue     *list.List

	wg *sync.WaitGroup

	ifResponse   bool
	responses    []interface{} // record response
	responsesMux sync.Mutex    // prevent concurrent write since no idea size of job
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

func (d *parallelDispatcher) Add(job func()) {
	d.AddWithResponse(func() interface{} {
		job()
		return nil
	})
}

func (d *parallelDispatcher) AddWithResponse(job func() interface{}) {
	d.wg.Add(1)

	if d.ifResponse {
		d.responsesMux.Lock()
		defer d.responsesMux.Unlock()
		d.responses = append(d.responses, nil)
	}

	d.mux.Lock()
	defer d.mux.Unlock()

	d.queue.PushBack(payload{job: job, seq: len(d.responses) - 1})
	d.queueCond.Signal()
}

func (d *parallelDispatcher) Wait() []interface{} {
	d.wg.Wait()
	d.Stop()
	return d.responses
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
	dispatcher *parallelDispatcher
	jobChannel chan payload
	quit       chan bool
}

func newParallelWorker(d *parallelDispatcher) parallelWorker {
	return parallelWorker{
		dispatcher: d,
		jobChannel: make(chan payload),
		quit:       make(chan bool),
	}
}

func (w parallelWorker) start() {
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

func (w parallelWorker) stop() {
	w.quit <- true
}
