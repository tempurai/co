package co_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempurai/co"
)

func TestAsyncMergedSequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		numbers := []int{1, 4, 5, 6, 7}
		aList := co.OfListWith(numbers...)

		numbers2 := []int{2, 4, 7, 0, 21}
		aList2 := co.OfListWith(numbers2...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) int {
			return v + 1
		})

		pList := co.NewAsyncMergedSequence[int](mList, aList2)

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := []int{2, 2, 5, 4, 6, 7, 7, 0, 8, 21}
			actual := []int{}
			for data := range pList.Iter() {
				actual = append(actual, data)
			}
			convey.So(actual, convey.ShouldResemble, expected)

		})
	})
}
