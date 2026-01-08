package pool_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempurai/co/ds/pool"
)

func TestWorkerPool(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		markers := make([]bool, 1000)

		p := pool.NewWorkerPool[int](10)

		for i := 0; i < 1000; i++ {
			func(idx int) {
				p.AddJob(func() int {
					markers[idx] = true
					return idx
				})
			}(i)
		}

		p.Wait()

		convey.Convey("Each markers should be marked", func() {
			for i := 0; i < 1000; i++ {
				convey.So(markers[i], convey.ShouldEqual, true)
			}
		})
	})
}
