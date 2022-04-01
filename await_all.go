package co

import (
	"log"
	"sync"
)

type actionAwait[R any] struct {
	*Action[*data[R]]

	list executableListIterator[R]
}

func (a *actionAwait[R]) run() {
	wg := sync.WaitGroup{}
	sData := NewSequenceableData[R]()

	for i := 0; a.list.hasNext(); i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			val, err := a.list.exeAt(i)

			log.Println("$$$$ DEBUG i val", i, val)
			sData.setAt(i, val, err)
		}(i)
	}

	wg.Wait()
	a.listenBulk(sData.GetAll())
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
