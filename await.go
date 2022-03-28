package co

func AwaitAll[R any](fns ...func() (R, error)) []*executor[R] {
	return All(NewConcurrent[R]().Add(fns...)).GetAll()
}

func AwaitRace[R any](fns ...func() (R, error)) R {
	return Race(NewConcurrent[R]().Add(fns...))
}

func AwaitAny[R any](fns ...func() (R, error)) R {
	return Any(NewConcurrent[R]().Add(fns...))
}
