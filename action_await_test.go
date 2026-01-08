//go:build !race

package co_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/tempurai/co"
)

func TestAwaitAll(t *testing.T) {
	convey.Convey("given a sequential tasks", t, func() {
		handlers := make([]func() (int, error), 0)
		for i := 0; i < 200; i++ {
			i := i
			handlers = append(handlers, func() (int, error) {
				return i + 1, nil
			})
		}

		convey.Convey("Run AwaitAll", func() {
			responses := co.AwaitAll(handlers...)

			convey.Convey("The responded value should be valid", func() {
				expected, actuals := []int{}, []int{}
				for i := 0; i < 200; i++ {
					expected = append(expected, i+1)
					actuals = append(actuals, responses[i].GetValue())
				}
				convey.So(expected, convey.ShouldResemble, actuals)
			})
		})
	})
}

func TestAwaitRace(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	convey.Convey("given a sequential tasks", t, func() {
		baseDelay := 10 * time.Millisecond
		handlers := make([]func() (int, error), 0)
		for i := 0; i < 10; i++ {
			i := i
			handlers = append(handlers, func() (int, error) {
				time.Sleep(baseDelay * time.Duration(i+1))
				return i + 1, nil
			})
		}

		convey.Convey("Run AwaitAll", func() {
			responses := co.AwaitRace(handlers...)

			convey.Convey("The responded value should be valid", func() {
				convey.So(responses, convey.ShouldEqual, 1)
			})
		})
	})
}

func TestAwaitAny(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	convey.Convey("given a sequential tasks", t, func() {
		baseDelay := 10 * time.Millisecond
		handlers := make([]func() (int, error), 0)
		for i := 0; i < 10; i++ {
			i := i

			err := fmt.Errorf("Determined value")
			if i > 3 {
				err = nil
			}

			handlers = append(handlers, func() (int, error) {
				time.Sleep(baseDelay * time.Duration(i+1))
				return i + 1, err
			})
		}

		convey.Convey("Run AwaitAll", func() {
			responses := co.AwaitAny(handlers...)

			convey.Convey("The responded value should be valid", func() {
				convey.So(responses.GetValue(), convey.ShouldEqual, 5)
			})
		})
	})
}
