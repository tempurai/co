package co_test

import (
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.tempura.ink/co"
)

func TestAsyncMapSequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		numbers := []int{1, 4, 5, 6, 7, 2, 2, 3, 4, 5, 12, 4, 2, 3, 43, 127, 37598, 34, 34, 123, 123}
		aList := co.OfListWith(numbers...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) string {
			return strconv.Itoa(v)
		})

		convey.Convey("expect resolved list to be identical with given values", func() {
			idx := 0
			for data := range mList.Iter() {
				convey.So(data, convey.ShouldEqual, strconv.Itoa(numbers[idx]))
				idx++
			}
		})
	})
}
