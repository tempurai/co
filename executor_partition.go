package co

type partitionBasic[R any] struct {
	its   []ExecutableIterator[R]
	fn    func(*SequenceableData[R])
	width int
}

func NewPartitionBasic[R any](its []ExecutableIterator[R], width int, fn func(*SequenceableData[R])) *partitionBasic[R] {
	return &partitionBasic[R]{
		its:   its,
		fn:    fn,
		width: width,
	}
}

func (p *partitionBasic[R]) ifHasNext() bool {
	for i := range p.its {
		if p.its[i].hasNext() {
			return true
		}
	}
	return false
}

func (p *partitionBasic[R]) exe() {
	seqData := NewSequenceableData[R]()

	for i := 0; p.ifHasNext(); i++ {
		for j := range p.its {
			data, err := p.its[j].exeNext()
			seqData.add(data, err)

			if seqData.len() == p.width || !p.ifHasNext() {
				p.fn(seqData)
				seqData = NewSequenceableData[R]()
			}
		}
	}
}

func Partition[R any](fn func(*SequenceableData[R]), width int, cos ...Concurrently[R]) {
	its := toConcurrentIterators(cos...)
	NewPartitionBasic(its, width, fn).exe()
}
