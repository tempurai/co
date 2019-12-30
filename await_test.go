package co_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tempera-shrimp/co"
	"testing"
)

func TestAwait(t *testing.T) {
	Convey("given a sequential tasks", t, func() {
		handlers := make([]func() interface{}, 0)
		for i := 0; i < 1000; i++ {
			i := i
			handlers = append(handlers, func() interface{} {
				return i + 1
			})
		}

		Convey("Run Await", func() {
			responses := co.Await(handlers...)

			Convey("The responded value should be valid", func() {
				for i := 0; i < 1000; i++ {
					So(responses[i], ShouldEqual, i+1)
				}
			})
		})
	})
}
