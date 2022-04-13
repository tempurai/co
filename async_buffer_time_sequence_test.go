package co_test

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestAsyncBufferTimeSequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		queued := []int{1, 4, 5, 6, 7, 2, 2, 3, 4, 5, 12, 4, 2, 3, 43, 127, 37598, 34, 34, 123, 123}
		sourceCh := make(chan int)

		oChannel := co.FromChan(sourceCh)
		bList := co.NewAsyncBufferTimeSequence[int](oChannel, time.Second)

		go func() {
			time.Sleep(time.Second)
			for i, val := range queued {
				sourceCh <- val
				time.Sleep(time.Millisecond * (100 + time.Duration(i)*10))
			}
			oChannel.Done()
		}()

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := [][]int{{1, 4, 5, 6, 7, 2, 2, 3}, {4, 5, 12, 4, 2}, {3, 43, 127, 37598, 34}, {34, 123, 123}}
			actual := [][]int{}
			for data := range bList.Iter() {
				actual = append(actual, data.GetValue())
			}
			convey.So(actual, convey.ShouldResemble, expected)
		})
	})
}
