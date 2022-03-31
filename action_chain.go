package co

type actionChain[R any] struct {
	*Action[*data[R]]
	cos []CoSequenceable[R]
}

func (a *actionChain[R]) run() {
	for i := 0; i < len(a.cos); i++ {
		a.listenBulk(Await(a.cos[i]).AsData().GetData())
	}
	a.done()
}

func Chain[R any](cos ...CoSequenceable[R]) *Action[*data[R]] {
	action := &actionChain[R]{
		Action: NewAction[*data[R]](),
		cos:    cos,
	}

	SafeGo(action.run)
	return action.Action
}
