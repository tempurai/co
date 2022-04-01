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
	emitFn       func(E)
	emitCh       []chan E
	emitDataList []E

	actionMode ActionMode
	closeChan  chan bool
	firstChan  chan bool

	rwmux sync.RWMutex
}

func NewAction[E any]() *Action[E] {
	return &Action[E]{
		emitCh:       make([]chan E, 0),
		emitDataList: make([]E, 0),
		closeChan:    make(chan bool),
		firstChan:    make(chan bool),
	}
}
func (a *Action[E]) Emitter() chan E {
	a.actionMode = ActionModeChan

	ch := make(chan E)
	a.emitCh = append(a.emitCh, ch)
	return ch
}

func (a *Action[E]) AsCallback(fn func(E)) *Action[E] {
	a.actionMode = ActionModeFn

	a.emitFn = fn
	return a
}

func (a *Action[E]) WaitData() *Action[E] {
	a.actionMode = ActionModeData
	return a
}

func (a *Action[E]) GetData() []E {
	a.wait()

	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	return a.emitDataList
}

func (a *Action[E]) PeakData() E {
	if len(a.emitDataList) == 0 {
		<-a.firstChan
	}

	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	if len(a.emitDataList) == 0 {
		return *new(E)
	}
	return a.emitDataList[0]
}

func (a *Action[E]) wait() {
	<-a.closeChan
}

func (a *Action[E]) emitData(e E) {
	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	for _, ch := range a.emitCh {
		SafeSend(ch, e)
	}
}

func (a *Action[E]) listenProgressive(e E) {
	switch a.actionMode {
	case ActionModeFn:
		a.emitFn(e)
	case ActionModeChan:
		a.emitData(e)
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
			a.emitData(e[i])
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

	sendFirstCh := len(a.emitDataList) == 0
	a.emitDataList = append(a.emitDataList, e...)

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
	for i := range a.emitCh {
		SafeClose(a.emitCh[i])
	}
	SafeClose(a.closeChan)
}

func MapAction[T1, T2 any](a1 *Action[T1], fn func(T1) T2) *Action[T2] {
	a2 := NewAction[T2]()
	a2.actionMode = a1.actionMode
	a2.closeChan = a1.closeChan

	a2.emitDataList = make([]T2, len(a1.emitDataList))
	for i := range a1.emitDataList {
		a2.emitDataList[i] = fn(a1.emitDataList[i])
	}
	return a2
}
