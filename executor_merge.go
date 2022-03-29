package co

type mergeBasic[R any] struct {
	cos   []*Concurrent[R]
	fn    func(*SequenceableData[R])
	width int
}

func NewMergeBasic[R any](cos []*Concurrent[R], width int, fn func(*SequenceableData[R])) *mergeBasic[R] {
	return &mergeBasic[R]{
		cos:   cos,
		fn:    fn,
		width: width,
	}
}

func (m *mergeBasic[R]) maxLen() int {
	l := 0
	for _, co := range m.cos {
		if l < co.len() {
			l = co.len()
		}
	}
	return l
}

func (m *mergeBasic[R]) exe() {
	mLen := m.maxLen()

	seqData := NewSequenceableData[R]()

	for i := 0; i < mLen; i++ {
		for j := range m.cos {
			data, err := m.cos[j].exeFnAt(i)

			seqData.resultAdd(data, err)

			if seqData.len() == m.width {
				m.fn(seqData)
				seqData = NewSequenceableData[R]()
			}
		}
	}
}

func Merge[R any](fn func(*SequenceableData[R]), width int, cos ...*Concurrent[R]) {
	NewMergeBasic(cos, width, fn).exe()
}
