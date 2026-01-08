package co_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempurai/co"
	"golang.org/x/exp/slices"
)

func checkCombineLatest[T comparable](c convey.C, a [][]T, l ...[]T) {
	cs := make([][]T, len(l))
	for _, v := range a {
		for i := range l {
			cs[i] = append(cs[i], v[i])
		}
	}
	for i := range l {
		c.So(slices.Compact(cs[i]), convey.ShouldResemble, l[i])
	}
}

func TestAsyncCombineLatest2Sequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func(c convey.C) {
		numbers := []int{1, 2, 3, 4, 5}
		aList := co.OfListWith(numbers...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) int {
			return v
		})

		numbers2 := []int{1, 2, 3, 4, 5}
		aList2 := co.OfListWith(numbers2...)

		pList := co.CombineLatest[int, int](mList, aList2)

		convey.Convey("expect resolved list to be identical with given values \n", func() {
			actual := [][]int{}
			for data := range pList.Iter() {
				actual = append(actual, []int{data.V1, data.V2})
			}
			convey.Printf("resulted list :: %+v \n", actual)
			checkCombineLatest(c, actual, numbers, numbers2)
		})
	})
}

func TestAsyncCombineLatest2SequenceWithDifferentLength(t *testing.T) {
	convey.Convey("given a sequential int", t, func(c convey.C) {
		numbers := []int{1, 2, 3, 4, 5, 6, 7, 8}
		aList := co.OfListWith(numbers...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) int {
			return v
		})

		numbers2 := []int{1, 2, 3, 4, 5}
		aList2 := co.OfListWith(numbers2...)

		pList := co.CombineLatest[int, int](mList, aList2)

		convey.Convey("expect resolved list to be identical with given values \n", func() {
			actual := [][]int{}
			for data := range pList.Iter() {
				actual = append(actual, []int{data.V1, data.V2})
			}
			convey.Printf("resulted list :: %+v \n", actual)
			checkCombineLatest(c, actual, numbers, numbers2)
		})
	})
}

func TestAsyncCombineLatest3Sequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func(c convey.C) {
		numbers := []int{1, 2, 3, 4, 5}
		aList := co.OfListWith(numbers...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) int {
			return v
		})

		numbers2 := []int{1, 2, 3, 4, 5}
		aList2 := co.OfListWith(numbers2...)

		numbers3 := []int{1, 2, 3, 4, 5}
		aList3 := co.OfListWith(numbers3...)

		pList := co.CombineLatest3[int, int, int](mList, aList2, aList3)

		convey.Convey("expect resolved list to be identical with given values \n", func() {
			actual := [][]int{}
			for data := range pList.Iter() {
				actual = append(actual, []int{data.V1, data.V2, data.V3})
			}
			convey.Printf("resulted list %+v \n", actual)
			checkCombineLatest(c, actual, numbers, numbers2, numbers3)
		})
	})
}
