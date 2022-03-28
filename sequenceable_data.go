package co

type SequenceableData[R any] struct {
	executorList[R]
}

func NewSequenceableData[R any]() *SequenceableData[R] {
	return &SequenceableData[R]{
		executorList: *NewExecutorList[R](),
	}
}

// ***************************************************************
// ********************** Chain Operations ***********************
// ***************************************************************
func (s *SequenceableData[R]) Filter(fn func(R) bool) *SequenceableData[R] {
	filteredSlice := make([]*executor[R], 0, len(s.executors))
	for i, executor := range s.executors {
		if executor.Error != nil {
			filteredSlice = append(filteredSlice, s.executors[i])
			continue
		}
		if fn(executor.Data) {
			filteredSlice = append(filteredSlice, s.executors[i])
		}
	}

	s.executors = filteredSlice
	return s
}

func (r *SequenceableData[R]) Peak() *executor[R] {
	if len(r.executors) == 0 {
		return nil
	}
	return r.executors[0]
}

func (r *SequenceableData[R]) GetAll() []*executor[R] {
	return r.executors
}

func (r *SequenceableData[R]) GetAllData() []R {
	data := make([]R, len(r.executors))
	for i := range r.executors {
		data = append(data, r.executors[i].Data)
	}
	return data
}
