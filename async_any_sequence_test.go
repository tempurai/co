package co_test

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"go.tempura.ink/co"
	"golang.org/x/exp/slices"
)

func TestAsyncAnySequence(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		queued1 := []int{1, 3, 5, 7, 9}
		sourceCh1 := make(chan int)
		oChannel1 := co.FromChan(sourceCh1)

		queued2 := []int{2, 4, 6, 8, 10}
		sourceCh2 := make(chan int)
		oChanne2 := co.FromChan(sourceCh2)

		bList := co.NewAsyncAnySequence[int](oChannel1, oChanne2)

		go func() {
			time.Sleep(time.Second)
			for i := range queued1 {
				sourceCh1 <- queued1[i]
				time.Sleep(time.Millisecond * 100)
				sourceCh2 <- queued2[i]
			}
			oChannel1.Complete()
			oChanne2.Complete()
		}()

		convey.Convey("expect resolved list to be identical with given values", func() {
			expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			actual := []int{}
			for data := range bList.Iter() {
				actual = append(actual, data)
			}
			slices.Sort(actual)
			convey.So(actual, convey.ShouldResemble, expected)
		})
	})
}
