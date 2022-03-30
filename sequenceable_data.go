package co

type SequenceableData[R any] struct {
	*determinedDataList[R]
}

func NewSequenceableData[R any]() *SequenceableData[R] {
	return &SequenceableData[R]{
		determinedDataList: NewDeterminedDataList[R](),
	}
}

func (r *SequenceableData[R]) Peak() *data[R] {
	if r.len() == 0 {
		return nil
	}
	return r.List.getAt(0)
}

func (r *SequenceableData[R]) GetAll() []*data[R] {
	return r.determinedDataList.list
}

func (r *SequenceableData[R]) GetAllData() []R {
	data := make([]R, r.len())
	for i := range r.list {
		data = append(data, r.List.getAt(i).value)
	}
	return data
}
