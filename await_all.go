package co

import (
	"sync"
)

type actionAwait[R any] struct {
	*Action[*data[R]]

	list executableListIterator[R]
}

func (a *actionAwait[R]) run() {
	wg := sync.WaitGroup{}
	dataList := NewList[*data[R]]()

	for i := 0; a.list.preflight(); i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			val, err := a.list.exeAt(i)

			dataList.setAt(i, NewDataWith(val, err))
		}(i)
	}

	wg.Wait()
	a.listenBulk(dataList.list)
	a.done()
}

func Await[R any](list *executablesList[R]) *Action[*data[R]] {
	action := &actionAwait[R]{
		Action: NewAction[*data[R]](),
		list:   list.Iterator(),
	}

	SafeGo(action.run)
	return action.Action
}

func AwaitAll[R any](fns ...func() (R, error)) []*data[R] {
	return Await(NewExecutablesList[R]().AddExecutable(fns...)).AsData().GetData()
}
