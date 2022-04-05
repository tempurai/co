package pool_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co/pool"
)

func TestQueue(t *testing.T) {
	convey.Convey("given a sequential int to enqueue", t, func() {
		q := pool.NewQueue[int]()
		l := 1000

		for i := 1; i <= l; i++ {
			q.Enqueue(i)
		}

		convey.Convey("On wait", func() {
			convey.Convey("Dequeue should be indentical to enqueued", func() {
				for i := 1; i <= l; i++ {
					v := q.Dequeue()
					convey.So(v, convey.ShouldEqual, i)
				}
			})
		})
	})
}
