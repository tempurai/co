package co_test

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestParallel(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		markers := make([]bool, 10000)

		p := co.NewParallel(10)
		for i := 0; i < 10000; i++ {
			func(idx int) {
				p.Add(func() {
					markers[idx] = true
				})
			}(i)
		}
		convey.Convey("On wait", func() {
			p.Wait()

			convey.Convey("Each markers should be marked", func() {
				for i := 0; i < 10000; i++ {
					convey.So(markers[i], convey.ShouldEqual, true)
				}
			})
		})
	})
}

func TestParallelWithResponse(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		p := co.NewParallelWithResponse[int](10)
		for i := 0; i < 10000; i++ {
			func(idx int) {
				p.AddWithResponse(func() int {
					return idx + 1
				})
			}(i)
		}

		convey.Convey("On wait", func() {
			vals := p.Wait()

			convey.Convey("Each response should be valid", func() {
				for i := 0; i < 10000; i++ {
					convey.So(vals[i], convey.ShouldEqual, i+1)
				}
			})
		})
	})
}

func TestParallelSeparatedAdd(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
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

		convey.Convey("On wait", func() {
			convey.Convey("Each markers should be marked", func() {
				for i := 0; i < 1000; i++ {
					convey.So(markers[i], convey.ShouldEqual, true)
				}
			})
		})
	})
}

func TestParallelHungerWait(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
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

		convey.Convey("On wait", func() {
			p.Wait()

			convey.Convey("Each markers should be marked", func() {
				for i := 0; i < 500; i++ {
					convey.So(markers[i], convey.ShouldEqual, true)
				}
			})
		})
	})
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
