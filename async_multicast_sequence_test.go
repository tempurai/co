package co_test

import (
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"tempura.ink/co"
)

func TestAsyncMulticastSequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		numbers := []int{1, 4, 5, 6, 7}
		aList := co.OfListWith(numbers...)
		mList := co.NewAsyncMapSequence[int](aList, func(v int) string {
			return strconv.Itoa(v)
		})
		cList := co.NewAsyncMulticastSequence[string](mList)
		cc1 := cList.Connect()
		cc2 := cList.Connect()

		convey.Convey("expect first resolved list to be identical with given values", func() {
			expected := []string{"1", "4", "5", "6", "7"}

			actual := []string{}
			for data := range cc1.Iter() {
				actual = append(actual, data)
			}
			convey.So(actual, convey.ShouldResemble, expected)
		})

		convey.Convey("expect second resolved list to be identical with given values", func() {
			expected := []string{"1", "4", "5", "6", "7"}

			actual := []string{}
			for data := range cc2.Iter() {
				actual = append(actual, data)
			}
			convey.So(actual, convey.ShouldResemble, expected)
		})
	})
}
