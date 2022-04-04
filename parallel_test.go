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

		p := co.NewParallel[bool](10)
		for i := 0; i < 10000; i++ {
			func(idx int) {
				p.Process(func() bool {
					markers[idx] = true
					return true
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
		p := co.NewParallel[int](10).SetPersistentData(true)
		for i := 0; i < 1000; i++ {
			func(idx int) {
				p.Process(func() int {
					return idx + 1
				})
			}(i)
		}

		convey.Convey("On wait", func() {
			vals := p.Wait().GetData()

			convey.Convey("Each response should be valid", func() {
				for i := 0; i < 1000; i++ {
					convey.So(vals[i], convey.ShouldEqual, i+1)
				}
			})
		})
	})
}

func TestParallelSeparatedAdd(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		markers := make([]bool, 10000)

		p := co.NewParallel[bool](10)
		for i := 0; i < 500; i++ {
			func(idx int) {
				p.Process(func() bool {
					markers[idx] = true
					return true
				})
			}(i)
		}

		// simulate sleep
		for i := 0; i < 2; i++ {
			time.Sleep(1 * time.Second)
		}

		for i := 500; i < 1000; i++ {
			func(idx int) {
				p.Process(func() bool {
					markers[idx] = true
					return true
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

		p := co.NewParallel[bool](10)
		for i := 0; i < 500; i++ {
			func(idx int) {
				p.Process(func() bool {
					markers[idx] = true
					return true
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
