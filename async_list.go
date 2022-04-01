package co

type AsyncList[R any] struct {
	*asyncSequence[R]

	*iterativeList[R]
}

func NewAsyncList[R any]() *AsyncList[R] {
	a := &AsyncList[R]{iterativeList: NewIterativeList[R]()}
	a.asyncSequence = NewAsyncSequence(a.Iterator())
	return a
}

func NewAsyncListWith[R any](val ...R) *AsyncList[R] {
	list := NewAsyncList[R]()
	return list.Add(val...)
}

func (it *AsyncList[R]) Add(e ...R) *AsyncList[R] {
	it.add(e...)
	return it
}

func (it *AsyncList[R]) Iterator() Iterator[R] {
	return asyncListIterator[R]{iterativeListIterator: it.iterativeList.Iterator()}
}

type asyncListIterator[R any] struct {
	*iterativeListIterator[R]
}
