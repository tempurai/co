package co

type mergeBasic[R any] struct {
	cos []*Concurrent[R]
	fn  func(R, error)

	cosIndex int
	exeIndex int
}

func NewMergeBasic[R any](cos []*Concurrent[R], fn func(R, error)) *mergeBasic[R] {
	return &mergeBasic[R]{
		cos: cos,
		fn:  fn,
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

	for i := 0; i < mLen; i++ {
		for j := range m.cos {
			data, err := m.cos[j].exeFnAt(i)
			m.fn(data, err)
		}
	}
}

func Merge[R any](fn func(R, error), cos ...*Concurrent[R]) {
	NewMergeBasic(cos, fn).exe()
}
