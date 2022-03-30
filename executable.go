package co

import (
	"log"
)

type executableFn[R any] func() (R, error)

type executable[R any] struct {
	fn       func() (R, error)
	executed bool
}

func NewExecutor[R any]() *executable[R] {
	return &executable[R]{}
}

func (r *executable[R]) setFn(fn func() (R, error)) *executable[R] {
	r.fn = fn
	return r
}

func (r *executable[R]) exe() (R, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("executable function error: %w", r)
		}
	}()
	return r.fn()
}

func (r *executable[R]) setExecuted(b bool) {
	r.executed = b
}

func (r *executable[R]) isExecuted() bool {
	return r.executed
}

type executablesList[R any] struct {
	*iterativeList[*executable[R]]
}

func NewExecutablesList[R any]() *executablesList[R] {
	return &executablesList[R]{
		iterativeList: NewIterativeList[*executable[R]](),
	}
}
