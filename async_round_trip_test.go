//go:build !race

package co_test

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestAsyncRoundTrip(t *testing.T) {
	convey.Convey("given a sequential int to enqueue", t, func(c convey.C) {
		source := []int{1, 4, 5, 6, 7, 2, 2, 3, 4, 5, 12, 4, 2, 3, 43, 127, 37598, 34, 34, 123, 123}
		actual := []int{}

		callbackFn := func(v int, _ error) {
			actual = append(actual, v)
		}
		aList := co.NewAsyncRoundTrip[int, int]()

		go func() {
			time.Sleep(time.Second)
			for i, val := range source {
				aList.SendItem(val, callbackFn)
				time.Sleep(time.Millisecond * (50 + time.Duration(i)*10))
			}
			time.Sleep(2 * time.Second)
			aList.Done()
		}()

		convey.Convey("expect resolved list to be identical with given values", func() {
			triggered := 0
			aList.Handle(func(i int) (int, error) {
				triggered++
				return i + 1, nil
			})

		})

		time.Sleep(time.Second * 1)
		expected := []int{2, 5, 6, 7, 8, 3, 3, 4, 5, 6, 13, 5, 3, 4, 44, 128, 37599, 35, 35, 124, 124}
		convey.So(actual, convey.ShouldResemble, expected)
	})
}
