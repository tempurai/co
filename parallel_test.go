package co_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func memoizeFib(n int) int {
	cache := make(map[int]int)
	result := make([]int, n)

	for i := 1; i <= n; i++ {
		result[i-1] = refinedExpensiveFib(i, cache)
	}

	return result[n-1]
}

func refinedExpensiveFib(n int, cache map[int]int) int {
	if n < 2 {
		cache[n] = n
		return n
	}
	if _, ok := cache[n-1]; !ok {
		cache[n-1] = refinedExpensiveFib(n-1, cache)
	}
	if _, ok := cache[n-2]; !ok {
		cache[n-2] = refinedExpensiveFib(n-2, cache)
	}
	return cache[n-1] + cache[n-2]
}

func TestParallel(t *testing.T) {
	Convey("given a sequential tasks", t, func() {
		markers := make([]bool, 10000)

		p := co.NewParallel(10)
		for i := 0; i < 10000; i++ {
			func(idx int) {
				p.Add(func() {
					markers[idx] = true
				})
			}(i)
		}
		Convey("On wait", func() {
			p.Wait()

			Convey("Each markers should be marked", func() {
				for i := 0; i < 10000; i++ {
					So(markers[i], ShouldEqual, true)
				}
			})
		})
	})
}

func TestParallelWithResponse(t *testing.T) {
	Convey("given a sequential tasks", t, func() {
		p := co.NewParallelWithResponse[int](10)
		for i := 0; i < 10000; i++ {
			func(idx int) {
				p.AddWithResponse(func() int {
					return idx + 1
				})
			}(i)
		}

		Convey("On wait", func() {
			vals := p.Wait()

			Convey("Each response should be valid", func() {
				for i := 0; i < 10000; i++ {
					So(vals[i], ShouldEqual, i+1)
				}
			})
		})
	})
}

func TestParallelSeparatedAdd(t *testing.T) {
	Convey("given a sequential tasks", t, func() {
		markers := make([]bool, 10000)

		p := co.NewParallel(10)
		for i := 0; i < 500; i++ {
			func(idx int) {
				p.Add(func() {
					markers[idx] = true
				})
			}(i)
		}

		// simulate sleep
		for i := 0; i < 2; i++ {
			time.Sleep(1 * time.Second)
		}

		for i := 500; i < 1000; i++ {
			func(idx int) {
				p.Add(func() {
					markers[idx] = true
				})
			}(i)
		}

		Convey("On wait", func() {
			Convey("Each markers should be marked", func() {
				for i := 0; i < 1000; i++ {
					So(markers[i], ShouldEqual, true)
				}
			})
		})
	})
}

func TestParallelHungerWait(t *testing.T) {
	Convey("given a sequential tasks", t, func() {
		markers := make([]bool, 10000)

		p := co.NewParallel(10)
		for i := 0; i < 500; i++ {
			func(idx int) {
				p.Add(func() {
					markers[idx] = true
				})
			}(i)
		}

		// simulate sleep
		for i := 0; i < 2; i++ {
			time.Sleep(1 * time.Second)
		}

		Convey("On wait", func() {
			p.Wait()

			Convey("Each markers should be marked", func() {
				for i := 0; i < 500; i++ {
					So(markers[i], ShouldEqual, true)
				}
			})
		})
	})
}

func BenchmarkWithoutParallel(b *testing.B) {
	for i := 1; i < b.N; i++ {
		memoizeFib(i)
	}
}

func BenchmarkParallel(b *testing.B) {
	p := co.NewParallel(15)

	for i := 1; i < b.N; i++ {
		func(idx int) {
			p.Add(func() {
				memoizeFib(idx)
			})
		}(i)
	}

	p.Wait()
}
