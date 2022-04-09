package co_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
	"golang.org/x/exp/slices"
)

func checkCombineLatest2[T comparable](c convey.C, l, r []T, a [][]T) {
	c1, c2 := []T{}, []T{}
	for _, v := range a {
		c1 = append(c1, v[0])
		c2 = append(c2, v[1])
	}
	c1 = slices.Compact(c1)
	c2 = slices.Compact(c2)

	c.So(c1, convey.ShouldResemble, l)
	c.So(c2, convey.ShouldResemble, r)
}

func checkCombineLatest3[T comparable](c convey.C, s1, s2, s3 []T, a [][]T) {
	c1, c2, c3 := []T{}, []T{}, []T{}
	for _, v := range a {
		c1 = append(c1, v[0])
		c2 = append(c2, v[1])
		c3 = append(c3, v[2])
	}
	c1 = slices.Compact(c1)
	c2 = slices.Compact(c2)
	c3 = slices.Compact(c3)

	c.So(c1, convey.ShouldResemble, s1)
	c.So(c2, convey.ShouldResemble, s2)
	c.So(c3, convey.ShouldResemble, s3)
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
			for data := range pList.Emit() {
				actual = append(actual, []int{data.GetValue().V1, data.GetValue().V2})
			}
			convey.Printf("resulted list :: %+v \n", actual)
			checkCombineLatest2(c, numbers, numbers2, actual)
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
			for data := range pList.Emit() {
				actual = append(actual, []int{data.GetValue().V1, data.GetValue().V2})
			}
			convey.Printf("resulted list :: %+v \n", actual)
			checkCombineLatest2(c, numbers, numbers2, actual)
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
			for data := range pList.Emit() {
				actual = append(actual, []int{data.GetValue().V1, data.GetValue().V2, data.GetValue().V3})
			}
			convey.Printf("resulted list %+v \n", actual)
			checkCombineLatest3(c, numbers, numbers2, numbers3, actual)
		})
	})
}
