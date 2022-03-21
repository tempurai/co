package co_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/tempera-shrimp/co"
)

func TestParallel(t *testing.T) {
	Convey("given a sequential tasks", t, func() {
		markers := make([]bool, 10000)

		p := co.NewParallel(10)
		for i := 0; i < 10000; i++ {
			i := i
			p.Add(func() {
				markers[i] = true
			})
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
			i := i
			p.AddWithResponse(func() int {
				return i + 1
			})
		}

		Convey("On wait", func() {
			vals := p.Wait()

			Convey("Each response should be valid", func() {
				for i := 0; i < 10000; i++ {
					So(vals[i], ShouldEqual, i + 1)
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
			i := i
			p.Add(func() {
				markers[i] = true
			})
		}

		// simulate sleep
		for i := 0; i < 10000000; i++ {

		}

		for i := 500; i < 1000; i++ {
			i := i
			p.Add(func() {
				markers[i] = true
			})
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
			i := i
			p.Add(func() {
				markers[i] = true
			})
		}

		// simulate sleep
		for i := 0; i < 10000000; i++ {

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