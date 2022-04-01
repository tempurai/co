package co

import (
	"sync"
)

type ActionMode int

const (
	ActionModeData ActionMode = iota
	ActionModeChan
	ActionModeFn
)

type Action[E any] struct {
	emitFn        func(E)
	emitCh        chan E
	emitCloseChan chan bool
	emitData      []E

	actionMode ActionMode
	closeChan  chan bool
	firstChan  chan bool

	rwmux sync.RWMutex
}

func NewAction[E any]() *Action[E] {
	return &Action[E]{
		emitCh:        make(chan E),
		emitData:      make([]E, 0),
		emitCloseChan: make(chan bool),
		closeChan:     make(chan bool),
		firstChan:     make(chan bool),
	}
}

func (a *Action[E]) AsChan() (chan E, chan bool) {
	a.actionMode = ActionModeChan
	return a.emitCh, a.emitCloseChan
}

func (a *Action[E]) Iterator() chan E {
	a.actionMode = ActionModeChan
	return a.emitCh
}

func (a *Action[E]) AsFn(fn func(E)) *Action[E] {
	a.actionMode = ActionModeFn
	a.emitFn = fn
	return a
}

func (a *Action[E]) AsData() *Action[E] {
	a.actionMode = ActionModeData
	return a
}

func (a *Action[E]) GetData() []E {
	a.wait()

	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	return a.emitData
}

func (a *Action[E]) PeakData() E {
	if len(a.emitData) == 0 {
		<-a.firstChan
	}

	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	if len(a.emitData) == 0 {
		return *new(E)
	}
	return a.emitData[0]
}

func (a *Action[E]) wait() {
	<-a.closeChan
}

func (a *Action[E]) listenProgressive(e E) {
	switch a.actionMode {
	case ActionModeFn:
		a.emitFn(e)
	case ActionModeChan:
		a.emitCh <- e
	case ActionModeData:
		a.appendToData(e)
	}
}

func (a *Action[E]) listenBulk(e []E) {
	switch a.actionMode {
	case ActionModeFn:
		for i := range e {
			a.emitFn(e[i])
		}
	case ActionModeChan:
		for i := range e {
			a.emitCh <- e[i]
		}
	case ActionModeData:
		a.appendToData(e...)
	}
}

func (a *Action[E]) appendToData(e ...E) {
	if len(e) == 0 {
		return
	}

	a.rwmux.Lock()
	defer a.rwmux.Unlock()

	sendFirstCh := len(a.emitData) == 0
	a.emitData = append(a.emitData, e...)

	if sendFirstCh {
		go func() {
			SafeSend(a.firstChan, true)
			SafeClose(a.firstChan)
		}()
	}
}

func (a *Action[E]) done() {
	switch a.actionMode {
	case ActionModeChan, ActionModeData:
		a.closeChan <- true
	}
	SafeClose(a.emitCh)
	SafeClose(a.emitCloseChan)
	SafeClose(a.closeChan)
}

func MapAction[T1, T2 any](a1 *Action[T1], fn func(T1) T2) *Action[T2] {
	a2 := NewAction[T2]()
	a2.actionMode = a1.actionMode
	a2.closeChan = a1.closeChan

	a2.emitData = make([]T2, len(a1.emitData))
	for i := range a1.emitData {
		a2.emitData[i] = fn(a1.emitData[i])
	}
	return a2
}
