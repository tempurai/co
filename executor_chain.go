package co

func Chain[R any](cos ...Concurrently[R]) *SequenceableData[R] {
	if len(cos) == 0 {
		return NewSequenceableData[R]()
	}

	seqData := NewSequenceableData[R]()
	for i := 0; i < len(cos); i++ {
		seqData.append(All(cos[i]).determinedDataList)
	}

	return seqData
}
