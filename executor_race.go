package co

func baseRace[R any](ignoreError bool, co *Concurrent[R]) R {
	respChan := make(chan R)
	aBool := &AtomicBool{}

	for i := range co.executors {
		go func(i int) {
			if aBool.Get() {
				return
			}
			if co.executors[i].isExecuted() {
				return
			}

			val, err := co.executors[i].exe()
			if err != nil && !ignoreError {
				return
			}

			SafeSend(respChan, val)
			aBool.Set(true)
		}(i)
	}

	val := <-respChan
	close(respChan)

	return val
}

func Race[R any](co *Concurrent[R]) R {
	return baseRace(true, co)
}

func Any[R any](co *Concurrent[R]) R {
	return baseRace(false, co)
}
