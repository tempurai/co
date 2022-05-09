package co_test

import (
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.tempura.ink/co"
)

func TestAsyncMulticastSequence(t *testing.T) {
	numbers := []int{1, 4, 5, 6, 7}
	aList := co.OfListWith(numbers...)
	mList := co.NewAsyncMapSequence[int](aList, func(v int) string {
		return strconv.Itoa(v)
	})
	cList := co.NewAsyncMulticastSequence[string](mList)
	cc1 := cList.Connect()
	cc2 := cList.Connect()

	convey.Convey("given a sequential int", t, func() {
		expected := []string{"1", "4", "5", "6", "7"}

		convey.Println("expect first resolved list to be identical with given values")

		actual1 := []string{}
		for data := range cc1.Iter() {
			actual1 = append(actual1, data)
		}
		convey.So(actual1, convey.ShouldResemble, expected)

		convey.Println("expect second resolved list to be identical with given values")
		actual2 := []string{}
		for data := range cc2.Iter() {
			actual2 = append(actual2, data)
		}
		convey.So(actual2, convey.ShouldResemble, expected)
	})
}
