package co_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/tempura-shrimp/co"
)

func TestAwaitAll(t *testing.T) {
	Convey("given a sequential tasks", t, func() {
		handlers := make([]func() (int, error), 0)
		for i := 0; i < 1000; i++ {
			i := i
			handlers = append(handlers, func() (int, error) {
				return i + 1, nil
			})
		}

		Convey("Run AwaitAll", func() {
			responses := co.AwaitAll(handlers...)

			Convey("The responded value should be valid", func() {
				for i := 0; i < 1000; i++ {
					So(responses[i].GetValue(), ShouldEqual, i+1)
				}
			})
		})
	})
}

func TestAwaitRace(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	Convey("given a sequential tasks", t, func() {
		handlers := make([]func() (int, error), 0)
		for i := 0; i < 100; i++ {
			i := i
			handlers = append(handlers, func() (int, error) {
				time.Sleep(time.Second * time.Duration(i+1))
				return i + 1, nil
			})
		}

		Convey("Run AwaitAll", func() {
			responses := co.AwaitRace(handlers...)

			Convey("The responded value should be valid", func() {
				So(responses, ShouldEqual, 1)
			})
		})
	})
}

func TestAwaitAny(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	Convey("given a sequential tasks", t, func() {
		handlers := make([]func() (int, error), 0)
		for i := 0; i < 100; i++ {
			i := i

			err := fmt.Errorf("Determined value")
			if i > 10 {
				err = nil
			}

			handlers = append(handlers, func() (int, error) {
				time.Sleep(time.Second * time.Duration(i+1))
				return i + 1, err
			})
		}

		Convey("Run AwaitAll", func() {
			responses := co.AwaitAny(handlers...)

			Convey("The responded value should be valid", func() {
				So(responses, ShouldEqual, 12)
			})
		})
	})
}

func BenchmarkAwaitAll(b *testing.B) {
	handlers := make([]func() (int, error), 0)
	for i := 1; i < b.N; i++ {
		func(idx int) {
			handlers = append(handlers, func() (int, error) {
				return memoizeFib(idx), nil
			})
		}(i)
	}

	co.AwaitAll(handlers...)
}
