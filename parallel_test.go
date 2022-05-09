package co_test

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"go.tempura.ink/co"
)

func TestParallel(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		size := 10000
		actual := make([]bool, size)
		expected := make([]bool, size)

		p := co.NewParallel[bool](10)
		for i := 0; i < size; i++ {
			func(idx int) {
				p.Process(func() bool {
					actual[idx] = true
					return true
				})
			}(i)
		}

		p.Wait()

		convey.Convey("Each markers should be marked", func() {
			for i := 0; i < size; i++ {
				expected[i] = true
			}
			convey.So(actual, convey.ShouldResemble, expected)
		})
	})
}

func TestParallelWithResponse(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		size := 10000
		p := co.NewParallel[int](10).SetPersistentData(true)
		for i := 0; i < size; i++ {
			func(idx int) {
				p.Process(func() int {
					return idx + 1
				})
			}(i)
		}

		actuals := p.Wait().GetData()
		expected := make([]int, 0)

		convey.Convey("Each response should be valid", func() {
			for i := 0; i < size; i++ {
				expected = append(expected, i+1)
			}
			convey.So(actuals, convey.ShouldResemble, expected)
		})
	})
}

func TestParallelSeparatedAdd(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		size1, size2 := 500, 1000
		actuals := make([]bool, size2)

		p := co.NewParallel[bool](10)
		for i := 0; i < size1; i++ {
			func(idx int) {
				p.Process(func() bool {
					actuals[idx] = true
					return true
				})
			}(i)
		}

		// simulate sleep
		for i := 0; i < 2; i++ {
			time.Sleep(1 * time.Second)
		}

		for i := size1; i < size2; i++ {
			func(idx int) {
				p.Process(func() bool {
					actuals[idx] = true
					return true
				})
			}(i)
		}

		p.Wait()
		expected := make([]bool, 0)

		convey.Convey("Each markers should be marked", func() {
			for i := 0; i < size2; i++ {
				expected = append(expected, true)

			}
			convey.So(actuals, convey.ShouldResemble, expected)
		})
	})
}

func TestParallelHungerWait(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		size := 500
		actuals := make([]bool, size)

		p := co.NewParallel[bool](10)
		for i := 0; i < size; i++ {
			func(idx int) {
				p.Process(func() bool {
					actuals[idx] = true
					return true
				})
			}(i)
		}

		// simulate sleep
		for i := 0; i < 2; i++ {
			time.Sleep(1 * time.Second)
		}

		p.Wait()

		expected := make([]bool, 0)
		convey.Convey("Each markers should be marked", func() {
			for i := 0; i < size; i++ {
				expected = append(expected, true)
			}
			convey.So(actuals, convey.ShouldResemble, expected)
		})
	})
}
