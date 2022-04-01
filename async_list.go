package co

type AsyncList[R any] struct {
	*iterativeList[R]
}

func NewAsyncList[R any]() *AsyncList[R] {
	return &AsyncList[R]{NewIterativeList[R]()}
}

func NewAsyncListWith[R any](val ...R) *AsyncList[R] {
	list := &AsyncList[R]{NewIterativeList[R]()}
	return list.Add(val...)
}

func (it *AsyncList[R]) Add(e ...R) *AsyncList[R] {
	it.add(e...)
	return it
}

func (it *AsyncList[R]) Iterator() asyncListIterator[R] {
	return asyncListIterator[R]{iterativeListIterator: it.iterativeList.Iterator()}
}

type asyncListIterator[R any] struct {
	*iterativeListIterator[R]
}
