package co_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

type vData struct{ v int }

func TestAsyncCompactedSequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		dataList := []*vData{{1}, nil, {4}, {3245}, {6}, nil, nil, {7}, nil, nil, nil, {32}, {4}}
		aList := co.OfListWith(dataList...)
		pList := co.NewAsyncCompactedSequence[*vData](aList)

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := []*vData{{1}, {4}, {3245}, {6}, {7}, {32}, {4}}
			actual := []*vData{}
			for data := range pList.Emitter() {
				actual = append(actual, data.GetValue())
			}
			convey.So(actual, convey.ShouldResemble, expected)

		})
	})
}

func TestAsyncCompactedSequenceStruct(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		dataList := []vData{{1}, {}, {4}, {3245}, {6}, {}, {}, {7}, {}, {}, {}, {32}, {4}}
		aList := co.OfListWith(dataList...)
		pList := co.NewAsyncCompactedSequence[vData](aList)

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := []vData{{1}, {4}, {3245}, {6}, {7}, {32}, {4}}
			actual := []vData{}
			for data := range pList.Emitter() {
				actual = append(actual, data.GetValue())
			}
			convey.So(actual, convey.ShouldResemble, expected)

		})
	})
}

func TestAsyncCompactedSequencePrimitiveTypes(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		dataList := []int{1, 23, 0, 5, 8, 0, 1, 3, 40, 0, 73, 12, 32345, 0, 123, 0, 123, 0, 0, 0, 0, 324, 1, 5, 6}
		aList := co.OfListWith(dataList...)
		pList := co.NewAsyncCompactedSequence[int](aList)

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := []int{1, 23, 5, 8, 1, 3, 40, 73, 12, 32345, 123, 123, 324, 1, 5, 6}
			actual := []int{}
			for data := range pList.Emitter() {
				actual = append(actual, data.GetValue())
			}
			convey.So(actual, convey.ShouldResemble, expected)

		})
	})
}
