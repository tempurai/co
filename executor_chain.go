package co

func Chain[R any](cos ...*Concurrent[R]) *SequenceableData[R] {
	if len(cos) == 0 {
		return NewSequenceableData[R]()
	}
	if len(cos) != 1 {
		for i := 1; i < len(cos); i++ {
			cos[0].Append(cos[i])
		}
	}
	return All(cos[0])
}
