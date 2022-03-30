package co

func AwaitAll[R any](fns ...func() (R, error)) []*data[R] {
	return All[R](NewConcurrent[R]().addExeFn(fns...)).asData().getData()
}

func AwaitRace[R any](fns ...func() (R, error)) R {
	return Race[R](NewConcurrent[R]().addExeFn(fns...)).asData().peakData()
}

func AwaitAny[R any](fns ...func() (R, error)) *data[R] {
	return Any[R](NewConcurrent[R]().addExeFn(fns...)).asData().peakData()
}
