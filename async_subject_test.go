package co_test

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"tempura.ink/co"
)

func TestAsyncSubject(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		expected := []int{1, 4, 5, 6, 7, 2, 2, 3, 4, 5, 12, 4, 2, 3, 43, 127, 37598, 34, 34, 123, 123}
		subject := co.NewAsyncSubject[int]()

		go func() {
			time.Sleep(time.Second)
			for i, val := range expected {
				subject.Next(val)
				time.Sleep(time.Millisecond * (100 + time.Duration(i)*10))
			}
			subject.Complete()
		}()

		convey.Convey("expect resolved list to be identical with given values", func() {
			actual, pre, intervals := []int{}, time.Now(), []time.Duration{}
			for data := range subject.Iter() {
				actual = append(actual, data)
				intervals = append(intervals, time.Since(pre))
				pre = time.Now()
			}

			convey.So(actual, convey.ShouldResemble, expected)

			intervals = intervals[1:]
			convey.Printf("We have a list of interval %+v\n", intervals)
			for i := range intervals {
				convey.So(intervals[i], convey.ShouldBeGreaterThan, time.Millisecond*(100+time.Duration(i)*10))
			}
		})
	})
}

func TestAsyncSubjectWait1Sec(t *testing.T) {
	convey.Convey("given a sequential int", t, func() {
		source := []int{1, 4, 5, 6, 7, 2, 2, 3, 4, 5, 12, 4, 2, 3, 43, 127, 37598, 34, 34, 123, 123}
		subject := co.NewAsyncSubject[int]()

		go func() {
			for i, val := range source {
				subject.Next(val)
				time.Sleep(time.Millisecond * (100 + time.Duration(i)*10))
			}
			subject.Complete()
		}()

		convey.Convey("expect resolved list to be identical with given values", func() {
			time.Sleep(time.Second)

			expected := []int{3, 4, 5, 12, 4, 2, 3, 43, 127, 37598, 34, 34, 123, 123}
			actual, pre, intervals := []int{}, time.Now(), []time.Duration{}
			for data := range subject.Iter() {
				actual = append(actual, data)
				intervals = append(intervals, time.Since(pre))
				pre = time.Now()
			}

			convey.So(actual, convey.ShouldResemble, expected)

			intervals = intervals[2:]
			convey.Printf("We have a list of interval %+v\n", intervals)
			for i := range intervals {
				convey.So(intervals[i], convey.ShouldBeGreaterThan, time.Millisecond*(100+time.Duration(i)*10))
			}
		})
	})
}
