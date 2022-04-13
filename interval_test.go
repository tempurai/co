package co_test

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestAsyncInterval(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		oChannel := co.Interval(100)

		convey.Convey("expect resolved list to be identical with given values", func() {
			counter, pre, intervals := 0, time.Now(), []int64{}
			for range oChannel.Iter() {
				intervals = append(intervals, time.Since(pre).Milliseconds())
				pre = time.Now()

				if counter++; counter == 20 {
					oChannel.Complete()
					break
				}
			}

			intervals = intervals[1:]
			convey.Printf("We have a list of interval %+v\n", intervals)
			for i := range intervals {
				convey.So(intervals[i], convey.ShouldBeBetweenOrEqual, 95, 105)
			}
			convey.So(len(intervals), convey.ShouldEqual, 19)
		})
	})
}
