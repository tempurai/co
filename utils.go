package co

import (
	"golang.org/x/exp/constraints"
)

func Copy[T any](v *T) *T {
	v2 := *v
	return &v2
}

func CastOrNil[T any](el any) T {
	if el == nil {
		return *new(T)
	}
	return el.(T)
}

func EvertGET[T constraints.Ordered](ele []T, target T) bool {
	for _, e := range ele {
		if e <= target {
			return false
		}
	}
	return true
}

func EvertET[T comparable](ele []T, target T) bool {
	for _, e := range ele {
		if e != target {
			return false
		}
	}
	return true
}

func AnyET[T comparable](ele []T, target T) bool {
	for _, e := range ele {
		if e == target {
			return true
		}
	}
	return false
}
