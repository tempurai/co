package co

type actionAnyResult struct {
	index int
	data  any
	err   error
}

type BulkResult[R any] struct {
	Value        R
	Err          error
	ReachesToEnd bool
}

type Type2[T1, T2 any] struct {
	V1 T1
	V2 T2
}

type Type3[T1, T2, T3 any] struct {
	V1 T1
	V2 T2
	V3 T3
}
