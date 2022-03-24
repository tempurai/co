package co_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

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
