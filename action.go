package co

type ActionMode int

const (
	ActionModeData ActionMode = iota
	ActionModeChan
	ActionModeFn
)

type Action[E any] struct {
	externalFn        func(E)
	externalCh        chan E
	externalCloseChan chan bool
	externalData      []E

	actionMode ActionMode
	closeChan  chan bool
}

func NewAction[E any]() *Action[E] {
	return &Action[E]{
		externalCh:        make(chan E),
		externalData:      make([]E, 0),
		externalCloseChan: make(chan bool),
		closeChan:         make(chan bool),
	}
}

func (a *Action[E]) asChan() (chan E, chan bool) {
	a.actionMode = ActionModeChan
	return a.externalCh, a.externalCloseChan
}

func (a *Action[E]) asFn(fn func(E)) *Action[E] {
	a.actionMode = ActionModeFn
	a.externalFn = fn
	return a
}

func (a *Action[E]) asData() *Action[E] {
	a.actionMode = ActionModeData
	return a
}

func (a *Action[E]) getData() []E {
	<-a.closeChan
	return a.externalData
}

func (a *Action[E]) listenProgressive(e E) {
	switch a.actionMode {
	case ActionModeFn:
		a.externalFn(e)
	case ActionModeChan:
		a.externalCh <- e
	case ActionModeData:
		a.externalData = append(a.externalData, e)
	}
}

func (a *Action[E]) listenBulk(e []E) {
	switch a.actionMode {
	case ActionModeFn:
		for i := range e {
			a.externalFn(e[i])
		}
	case ActionModeChan:
		for i := range e {
			a.externalCh <- e[i]
		}
	case ActionModeData:
		a.externalData = append(a.externalData, e...)
	}
}

func (a *Action[E]) done() {
	switch a.actionMode {
	case ActionModeChan, ActionModeData:
		a.closeChan <- true
	}
	SafeClose(a.externalCh)
	SafeClose(a.externalCloseChan)
	SafeClose(a.closeChan)
}
