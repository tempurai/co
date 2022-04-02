package co

import (
	"sync"
)

type Action[E any] struct {
	emitChs      []chan E
	bufferedData []E

	ifWaitData bool
	firstChan  chan bool
	lastChan   chan bool

	rwmux sync.RWMutex
}

func NewAction[E any]() *Action[E] {
	return &Action[E]{
		emitChs:      make([]chan E, 0),
		bufferedData: make([]E, 0),
		firstChan:    make(chan bool),
		lastChan:     make(chan bool),
		ifWaitData:   true,
	}
}
func (a *Action[E]) Emitter() chan E {
	ch := make(chan E)
	a.emitChs = append(a.emitChs, ch)
	return ch
}

func (a *Action[E]) DiscardData() *Action[E] {
	a.ifWaitData = false
	return a
}

func (a *Action[E]) wait() {
	<-a.lastChan
}

func (a *Action[E]) GetData() []E {
	a.wait()

	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	return a.bufferedData
}

func (a *Action[E]) PeakData() E {
	if len(a.bufferedData) == 0 {
		<-a.firstChan
	}

	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	if len(a.bufferedData) == 0 {
		return *new(E)
	}

	return a.bufferedData[0]
}

func (a *Action[E]) listen(el ...E) {
	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	for _, e := range el {
		for _, ch := range a.emitChs {
			SafeSend(ch, e)
		}
	}

	sendFirstCh := len(a.bufferedData) == 0
	if a.ifWaitData {
		a.bufferedData = append(a.bufferedData, el...)
	}
	if sendFirstCh {
		go func() {
			SafeSend(a.firstChan, true)
			SafeClose(a.firstChan)
		}()
	}
}

func (a *Action[E]) done() {
	for i := range a.emitChs {
		SafeClose(a.emitChs[i])
	}
	SafeSend(a.lastChan, true)
	SafeClose(a.lastChan)
}

func MapAction[T1, T2 any](a1 *Action[T1], fn func(T1) T2) *Action[T2] {
	a2 := NewAction[T2]()
	a2.firstChan = a1.firstChan
	a2.lastChan = a1.lastChan

	a2.bufferedData = make([]T2, len(a1.bufferedData))
	for i := range a1.bufferedData {
		a2.bufferedData[i] = fn(a1.bufferedData[i])
	}
	return a2
}
