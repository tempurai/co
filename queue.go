// source: https://colobu.com/2020/08/14/lock-free-queue-in-go/
package co

import (
	"sync/atomic"
	"unsafe"
)

// Queue is a lock-free unbounded queue.
type Queue[K any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer

	len int32
}

type node[K any] struct {
	value K
	next  unsafe.Pointer
}

// NewQueue returns an empty queue.
func NewQueue[K any]() *Queue[K] {
	n := unsafe.Pointer(&node[K]{})
	return &Queue[K]{head: n, tail: n}
}

func (q *Queue[K]) Len() int {
	return int(q.len)
}

// Enqueue puts the given value v at the tail of the queue.
func (q *Queue[K]) Enqueue(v K) {
	n := &node[K]{value: v}
	for {
		tail := load[K](&q.tail)
		next := load[K](&tail.next)
		if tail == load[K](&q.tail) { // are tail and next consistent?
			if next == nil {
				if cas(&tail.next, next, n) {
					atomic.AddInt32(&q.len, 1)
					cas(&q.tail, tail, n) // Enqueue is done.  try to swing tail to the inserted node
					return
				}
			} else { // tail was not pointing to the last node
				// try to swing Tail to the next node
				cas(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *Queue[K]) Dequeue() K {
	for {
		head := load[K](&q.head)
		tail := load[K](&q.tail)
		next := load[K](&head.next)
		if head == load[K](&q.head) { // are head, tail, and next consistent?
			if head == tail { // is queue empty or tail falling behind?
				if next == nil { // is queue empty?
					return *new(K)
				}
				// tail is falling behind.  try to advance it
				cas(&q.tail, tail, next)
			} else {
				// read value before CAS otherwise another dequeue might free the next node
				v := next.value
				if cas(&q.head, head, next) {
					atomic.AddInt32(&q.len, -1)
					return v // Dequeue is done.  return
				}
			}
		}
	}
}

func load[K any](p *unsafe.Pointer) (n *node[K]) {
	return (*node[K])(atomic.LoadPointer(p))
}

func cas[K any](p *unsafe.Pointer, old, new *node[K]) (ok bool) {
	return atomic.CompareAndSwapPointer(
		p, unsafe.Pointer(old), unsafe.Pointer(new))
}
