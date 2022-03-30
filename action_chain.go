package co

type ActionChain[R any] struct {
	*Action[*data[R]]
	cos []Concurrently[R]
}

func (a *ActionChain[R]) run() {
	for i := 0; i < len(a.cos); i++ {
		a.listenBulk(Await(a.cos[i]).asData().getData())
	}
	a.done()
}

func Chain[R any](cos ...Concurrently[R]) *Action[*data[R]] {
	action := &ActionChain[R]{
		Action: NewAction[*data[R]](),
		cos:    cos,
	}

	SafeGo(action.run)
	return action.Action
}
