package co_test

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestAsyncChannel(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		expected := []int{1, 4, 5, 6, 7, 2, 2, 3, 4, 5, 12, 4, 2, 3, 43, 127, 37598, 34, 34, 123, 123}
		sourceCh := make(chan int)

		oChannel := co.FromChan(sourceCh)

		go func() {
			time.Sleep(time.Second)
			for i, val := range expected {
				sourceCh <- val
				time.Sleep(time.Millisecond * (100 + time.Duration(i)*10))
			}
			oChannel.Done()
		}()

		convey.Convey("expect resolved list to be identical with given values", func() {
			actual, pre, intervals := []int{}, time.Now(), []time.Duration{}
			for data := range oChannel.Emitter() {
				actual = append(actual, data.GetValue())
				intervals = append(intervals, time.Now().Sub(pre))
				pre = time.Now()
			}

			convey.So(actual, convey.ShouldResemble, expected)

			intervals = intervals[1:]
			convey.Printf("We have a list of interval %+v\n", intervals)
			for i := range intervals {
				convey.So(intervals[i], convey.ShouldBeGreaterThan, time.Millisecond*(100+time.Duration(i)*10))
			}
		})
	})
}
