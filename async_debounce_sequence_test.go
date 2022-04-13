package co_test

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestAsyncDebounceSequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		queued := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}
		sourceCh := make(chan int)

		oChannel := co.FromChan(sourceCh)
		dList := co.NewAsyncDebounceSequence[int](oChannel, time.Second)

		go func() {
			time.Sleep(time.Second)
			for _, val := range queued {
				sourceCh <- val
				time.Sleep(time.Millisecond * 100)
			}
			oChannel.Complete()
		}()

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := []int{1, 11, 21}
			actual := []int{}
			for data := range dList.Iter() {
				actual = append(actual, data)
			}
			convey.So(actual, convey.ShouldResemble, expected)
		})
	})
}

func TestAsyncDebounceSequenceTolerance(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		queued := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}
		sourceCh := make(chan int)

		oChannel := co.FromChan(sourceCh)
		dList := co.NewAsyncDebounceSequence[int](oChannel, time.Second).SetTolerance(time.Millisecond * 200)

		go func() {
			time.Sleep(time.Second)
			for _, val := range queued {
				sourceCh <- val
				time.Sleep(time.Millisecond * 100)
			}
			oChannel.Complete()
		}()

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := []int{1, 2, 3, 13, 14, 15, 25}
			actual := []int{}
			for data := range dList.Iter() {
				actual = append(actual, data)
			}
			convey.So(actual, convey.ShouldResemble, expected)
		})
	})
}
