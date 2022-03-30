package co

func baseRace[R any](ignoreError bool, co Concurrently[R]) R {
	it := co.Iterator()

	respChan := make(chan R)
	aBool := &AtomicBool{}

	for i := 0; it.hasNext(); i++ {
		go func(i int, fn executableFn[R]) {
			if aBool.Get() {
				return
			}

			val, err := fn()
			if err != nil && !ignoreError {
				return
			}

			SafeSend(respChan, val)
			aBool.Set(true)
		}(i, it.exeFn())
	}

	val := <-respChan
	close(respChan)

	return val
}

func Race[R any](it Concurrently[R]) R {
	return baseRace(true, it)
}

func Any[R any](it Concurrently[R]) R {
	return baseRace(false, it)
}
