package co

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

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

func EvertGET[T Ordered](ele []T, target T) bool {
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
