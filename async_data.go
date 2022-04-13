package co

import (
	"sync/atomic"
)

type asyncStatus = int32

const (
	asyncStatusIdle asyncStatus = iota
	asyncStatusPending
	asyncStatusDone
)

type AsyncData[R any] struct {
	status asyncStatus
	value  R
}

func NewAsyncData[R any](val R) *AsyncData[R] {
	return &AsyncData[R]{
		status: asyncStatusIdle,
		value:  val,
	}
}

func (as *AsyncData[R]) TransiteTo(status asyncStatus) *AsyncData[R] {
	atomic.StoreInt32(&as.status, status)
	return as
}

func (as *AsyncData[R]) IsIdel() bool {
	return atomic.LoadInt32(&as.status) == asyncStatusIdle
}

func (as *AsyncData[R]) IsPending() bool {
	return atomic.LoadInt32(&as.status) == asyncStatusPending
}

func (as *AsyncData[R]) IsComplete() bool {
	return atomic.LoadInt32(&as.status) == asyncStatusDone
}
