package co_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestAsyncPartitionSequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		numbers := []int{1, 4, 5, 6, 7}
		aList := co.OfListWith(numbers...)

		numbers1 := []int{2, 3, 6, 7, 8}
		aList2 := co.OfListWith(numbers1...)

		pList := co.NewAsyncPartitionSequence[int](2, aList, aList2)

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := [][]int{{1, 2}, {4, 3}, {5, 6}, {6, 7}, {7, 8}}

			actual := [][]int{}
			for data := range pList.Iter() {
				actual = append(actual, data)
			}
			convey.So(actual, convey.ShouldResemble, expected)

		})
	})
}

func TestAsyncPartitionSequence3(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		numbers := []int{1, 4, 5, 6, 7}
		aList := co.OfListWith(numbers...)

		numbers1 := []int{2, 3, 6, 7, 8}
		aList2 := co.OfListWith(numbers1...)

		pList := co.NewAsyncPartitionSequence[int](3, aList, aList2)

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := [][]int{{1, 2, 4}, {3, 5, 6}, {6, 7, 7}, {8}}

			actual := [][]int{}
			for data := range pList.Iter() {
				actual = append(actual, data)
			}
			convey.So(actual, convey.ShouldResemble, expected)

		})
	})
}
