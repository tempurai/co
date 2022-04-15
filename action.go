package co

import (
	"sync"

	co_sync "github.com/tempura-shrimp/co/internal/sync"
)

type Action[E any] struct {
	emitChs      []chan E
	bufferedData *List[E]

	ifWaitData bool
	firstChan  chan bool
	lastChan   chan bool

	rwmux sync.RWMutex
}

func NewAction[E any]() *Action[E] {
	return &Action[E]{
		emitChs:      make([]chan E, 0),
		bufferedData: NewList[E](),
		firstChan:    make(chan bool),
		lastChan:     make(chan bool),
		ifWaitData:   true,
	}
}
func (a *Action[E]) Iter() chan E {
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

	return a.bufferedData.list
}

func (a *Action[E]) PeakData() E {
	if a.bufferedData.len() == 0 {
		<-a.firstChan
	}

	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	if a.bufferedData.len() == 0 {
		return *new(E)
	}

	return a.bufferedData.getAt(0)
}

func (a *Action[E]) listen(el ...E) {
	a.rwmux.RLock()
	defer a.rwmux.RUnlock()

	for _, e := range el {
		for _, ch := range a.emitChs {
			co_sync.SafeNSend(ch, e)
		}
	}

	sendFirstCh := a.bufferedData.len() == 0
	if a.ifWaitData {
		a.bufferedData.add(el...)
	}
	if sendFirstCh {
		go func() {
			co_sync.SafeSend(a.firstChan, true)
			co_sync.SafeClose(a.firstChan)
		}()
	}
}

func (a *Action[E]) done() {
	for i := range a.emitChs {
		co_sync.SafeClose(a.emitChs[i])
	}
	co_sync.SafeSend(a.lastChan, true)
	co_sync.SafeClose(a.lastChan)
}

func MapAction[T1, T2 any](a1 *Action[T1], fn func(T1) T2) *Action[T2] {
	a2 := NewAction[T2]()
	a2.firstChan = a1.firstChan
	a2.lastChan = a1.lastChan

	a2.bufferedData = NewList[T2]()
	a2.bufferedData.resizeTo(a1.bufferedData.len())
	for i := range a1.bufferedData.list {
		a2.bufferedData.list[i] = fn(a1.bufferedData.list[i])
	}
	return a2
}
