package co

func PairWise[R any](co Concurrently[R]) []*SequenceableData[R] {
	return TupleWise[R](co, 2)
}

func TupleWise[R any](co Concurrently[R], width int) []*SequenceableData[R] {
	seqs := make([]*SequenceableData[R], 0, width)
	return seqs
}
