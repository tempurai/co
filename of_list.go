package co

type AsyncList[R any] struct {
	*asyncSequence[R]

	*iterativeList[R]
}

func OfList[R any]() *AsyncList[R] {
	a := &AsyncList[R]{iterativeList: NewIterativeList[R]()}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func OfListWith[R any](val ...R) *AsyncList[R] {
	list := OfList[R]()
	return list.Add(val...)
}

func (it *AsyncList[R]) Add(e ...R) *AsyncList[R] {
	it.add(e...)
	return it
}

func (it *AsyncList[R]) iterator() Iterator[R] {
	return asyncListIterator[R]{iterativeListIterator: it.iterativeList.iterator()}
}

type asyncListIterator[R any] struct {
	*iterativeListIterator[R]
}
