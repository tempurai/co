package co_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func checkZip[T comparable](c convey.C, a [][]T, l ...[]T) {
	cs := make([][]T, len(l))
	for _, v := range a {
		for i := range l {
			cs[i] = append(cs[i], v[i])
		}
	}

	for i := range l {
		c.So(cs[i], convey.ShouldResemble, l[i])
	}
}

func TestAsyncZip2Sequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func(c convey.C) {
		list1 := []int{1, 2, 3, 4, 5}
		aList := co.OfListWith(list1...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) int {
			return v
		})

		list2 := []int{1, 2, 3, 4, 5}
		aList2 := co.OfListWith(list2...)

		pList := co.Zip[int, int](mList, aList2)

		convey.Convey("expect resolved list to be identical with given values \n", func() {
			actual := [][]int{}
			for data := range pList.Emit() {
				actual = append(actual, []int{data.GetValue().V1, data.GetValue().V2})
			}
			convey.Printf("resulted list :: %+v \n", actual)
			checkZip(c, actual, list1, list2)
		})
	})
}

func TestAsyncZip2SequenceWithDifferentLength(t *testing.T) {
	convey.Convey("given a sequential int", t, func(c convey.C) {
		list1 := []int{1, 2, 3, 4, 5, 6, 7, 8}
		aList := co.OfListWith(list1...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) int {
			return v + 1
		})

		list2 := []int{1, 2, 3, 4, 5}
		aList2 := co.OfListWith(list2...)

		pList := co.Zip[int, int](mList, aList2)

		convey.Convey("expect resolved list to be identical with given values \n", func() {
			expected1 := []int{2, 3, 4, 5, 6, 7, 8, 9}
			expected2 := []int{1, 2, 3, 4, 5, 5, 5, 5}

			actual := [][]int{}
			for data := range pList.Emit() {
				actual = append(actual, []int{data.GetValue().V1, data.GetValue().V2})
			}
			convey.Printf("resulted list :: %+v \n", actual)
			checkZip(c, actual, expected1, expected2)
		})
	})
}

func TestAsyncZip3Sequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func(c convey.C) {
		list1 := []int{1, 2, 3, 4, 5}
		aList := co.OfListWith(list1...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) int {
			return v + 1
		})

		list2 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		aList2 := co.OfListWith(list2...)

		list3 := []int{1, 2, 3, 4, 5, 6, 7}
		aList3 := co.OfListWith(list3...)

		pList := co.Zip3[int, int, int](mList, aList2, aList3)

		convey.Convey("expect resolved list to be identical with given values \n", func() {
			expected1 := []int{2, 3, 4, 5, 6, 6, 6, 6, 6}
			expected2 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
			expected3 := []int{1, 2, 3, 4, 5, 6, 7, 7, 7}

			actual := [][]int{}
			for data := range pList.Emit() {
				actual = append(actual, []int{data.GetValue().V1, data.GetValue().V2, data.GetValue().V3})
			}
			convey.Printf("resulted list %+v \n", actual)
			checkZip(c, actual, expected1, expected2, expected3)
		})
	})
}
