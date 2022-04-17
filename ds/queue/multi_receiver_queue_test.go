package queue_test

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"tempura.ink/co/ds/queue"
)

func TestMultiReceiverQueue(t *testing.T) {
	convey.Convey("given a sequential int to enqueue", t, func() {
		q := queue.NewMultiReceiverQueue[int]().Receiver()
		l := 1000

		expected := make([]int, 0)
		for i := 1; i <= l; i++ {
			expected = append(expected, i)
			q.Enqueue(i)
		}

		convey.Convey("On wait", func() {
			convey.Convey("Dequeue should be indentical to enqueued", func() {
				actual := make([]int, 0)
				for i := 1; i <= l; i++ {
					v := q.Dequeue()
					actual = append(actual, v)

				}
				convey.So(actual, convey.ShouldResemble, expected)
			})
		})
	})
}

func TestMultiReceiverQueueWith2Receiver(t *testing.T) {
	convey.Convey("given a sequential int to enqueue", t, func() {
		q := queue.NewMultiReceiverQueue[int]()
		r1 := q.Receiver()
		r2 := q.Receiver()
		l := 5000

		convey.Convey("Queue should be empty at initial", func() {
			convey.So(r1.IsEmpty(), convey.ShouldEqual, true)
			convey.So(r2.IsEmpty(), convey.ShouldEqual, true)
		})

		expected := make([]int, 0)
		for i := 0; i < l; i++ {
			expected = append(expected, i)
			q.Enqueue(i)
		}

		convey.Convey("On wait", func() {
			convey.Convey("Dequeue 1 should be indentical to enqueued", func() {
				convey.So(r1.IsEmpty(), convey.ShouldEqual, false)
				actual := make([]int, 0)
				for i := 0; i < l; i++ {
					v := r1.Dequeue()
					actual = append(actual, v)

				}
				convey.So(actual, convey.ShouldResemble, expected)
				convey.So(r1.IsEmpty(), convey.ShouldEqual, true)
			})
			convey.Convey("Dequeue 2 should be indentical to enqueued", func() {
				convey.So(r2.IsEmpty(), convey.ShouldEqual, false)
				actual := make([]int, 0)
				for i := 0; i < l; i++ {
					v := r2.Dequeue()
					actual = append(actual, v)

				}
				convey.So(actual, convey.ShouldResemble, expected)
				convey.So(r2.IsEmpty(), convey.ShouldEqual, true)
			})
		})
	})
}

func TestMultiReceiverQueueWith2ReceiverConcurrently(t *testing.T) {
	convey.Convey("given a sequential int to enqueue", t, func(c convey.C) {
		q := queue.NewMultiReceiverQueue[int]()
		rl, l := 50, 10000

		r := make([]*queue.QueueReceiver[int], 0)
		for i := 0; i < rl; i++ {
			r = append(r, q.Receiver())
		}

		convey.Convey("Queue should be empty at initial", func() {
			for i := 0; i < rl; i++ {
				convey.So(r[i].IsEmpty(), convey.ShouldEqual, true)
			}
		})

		expected := make([]int, 0)
		for i := 1; i <= l; i++ {
			expected = append(expected, i)
			q.Enqueue(i)
		}

		type result struct {
			i int
			v []int
		}

		actualCh := make(chan result)
		actualCount := (int32)(0)

		for i := 0; i < rl; i++ {
			go func(i int, re *queue.QueueReceiver[int]) {
				actual := make([]int, 0)
				for i := 1; i <= l; i++ {
					actual = append(actual, re.Dequeue())
				}
				actualCh <- result{i, actual}

				if atomic.AddInt32(&actualCount, 1) == int32(rl) {
					close(actualCh)
				}
			}(i, r[i])
		}

		for val := range actualCh {
			c.Convey(fmt.Sprintf("Dequeue %d should be indentical to enqueued", val.i+1), func() {
				c.So(val.v, convey.ShouldResemble, expected)
				c.So(r[val.i].IsEmpty(), convey.ShouldEqual, true)
			})
		}
	})
}
