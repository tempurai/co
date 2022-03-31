package co

func AwaitAll[R any](fns ...func() (R, error)) []*data[R] {
	return Await[R](NewCoExecutable[R]().AddFn(fns...)).AsData().GetData()
}

func AwaitRace[R any](fns ...func() (R, error)) R {
	return Race[R](NewCoExecutable[R]().AddFn(fns...)).AsData().PeakData()
}

func AwaitAny[R any](fns ...func() (R, error)) *data[R] {
	return Any[R](NewCoExecutable[R]().AddFn(fns...)).AsData().PeakData()
}
