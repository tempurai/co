package co_test

import (
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestAsyncPairwiseSequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		numbers := []int{1, 4, 5, 6, 7}
		aList := co.OfListWith(numbers...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) string {
			return strconv.Itoa(v)
		})
		pList := co.NewAsyncPairwiseSequence[string](mList)

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := [][]string{{"1", "4"}, {"4", "5"}, {"5", "6"}, {"6", "7"}}

			actual := [][]string{}
			for data := range pList.Iter() {
				actual = append(actual, data.GetValue())
			}
			convey.So(actual, convey.ShouldResemble, expected)

		})
	})
}
