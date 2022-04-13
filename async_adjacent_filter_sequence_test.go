package co_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestAsyncAdjacentFilterSequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		numbers := []int{1, 4, 5, 6, 7, 3, 5, 6, 2, 4, 6, 7, 8, 9, 3}
		aList := co.OfListWith(numbers...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) int {
			return v
		})
		pList := co.NewAsyncAdjacentFilterSequence[int](mList, func(r1, r2 int) bool {
			return r1 < r2
		})

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := []int{1, 4, 5, 6, 7, 8, 9}
			actual := []int{}
			for data := range pList.Iter() {
				actual = append(actual, data)
			}
			convey.So(actual, convey.ShouldResemble, expected)

		})
	})
}
