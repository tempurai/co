package co

func TupleWise[R any](co *Concurrent[R], width int) []*SequenceableData[R] {
	seqs := make([]*SequenceableData[R], 0, co.len()/width+1)

	Merge(func(seq *SequenceableData[R]) {
		seqs = append(seqs, seq)
	}, width, co)

	return seqs
}
