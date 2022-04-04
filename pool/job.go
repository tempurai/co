package pool

type job[K any] struct {
	fn  func() K
	seq uint64
}

type jobDone[K any] struct {
	val       K
	seq       uint64
	workerRef any
}
