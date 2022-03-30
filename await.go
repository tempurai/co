package co

func AwaitAll[R any](fns ...func() (R, error)) []*data[R] {
	return All[R](NewConcurrent[R]().addExeFn(fns...)).GetAll()
}

func AwaitRace[R any](fns ...func() (R, error)) R {
	return Race[R](NewConcurrent[R]().addExeFn(fns...))
}

func AwaitAny[R any](fns ...func() (R, error)) R {
	return Any[R](NewConcurrent[R]().addExeFn(fns...))
}
