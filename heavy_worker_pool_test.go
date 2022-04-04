package co_test

import (
	"testing"

	"github.com/Jeffail/tunny"
	"github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestWorkerPool(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		markers := make([]bool, 100)

		p := co.NewHeavyWorkerPool[any](10)

		for i := 0; i < 1000; i++ {
			func(idx int) {
				p.AddJob(func() any {
					markers[idx] = true
					return nil
				})
			}(i)
		}
		convey.Convey("On wait", func() {
			p.Wait()

			convey.Convey("Each markers should be marked", func() {
				for i := 0; i < 1000; i++ {
					convey.So(markers[i], convey.ShouldEqual, true)
				}
			})
		})
	})
}

func BenchmarkWorkPoolWithFib(b *testing.B) {
	p := co.NewHeavyWorkerPool[int](256)

	b.ResetTimer()
	for i := 1; i < b.N; i++ {
		func(idx int) {
			p.AddJob(func() int {
				return memoizeFib(idx)
			})
		}(i)
	}

	p.Wait()
}

func BenchmarkTunnyWithFib(b *testing.B) {
	pool := tunny.NewFunc(256, func(payload interface{}) interface{} {
		return memoizeFib(payload.(int))
	})
	defer pool.Close()

	b.ResetTimer()
	for i := 1; i < b.N; i++ {
		func(idx int) {
			pool.Process(idx)
		}(i)
	}
}
