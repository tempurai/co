package co

import (
	"log"
)

type executable[R any] struct {
	fn       func() (R, error)
	executed bool
}

func NewExecutor[R any]() *executable[R] {
	return &executable[R]{}
}

func (e *executable[R]) setFn(fn func() (R, error)) *executable[R] {
	e.fn = fn
	return e
}

func (e *executable[R]) exe() (R, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("executable function error: %w", r)
		}
	}()
	return e.fn()
}

func (e *executable[R]) setExecuted(b bool) {
	e.executed = b
}

func (e *executable[R]) isExecuted() bool {
	return e.executed
}

type executablesList[R any] struct {
	*iterativeList[*executable[R]]
}

func NewExecutablesList[R any]() *executablesList[R] {
	return &executablesList[R]{
		iterativeList: NewIterativeList[*executable[R]](),
	}
}
