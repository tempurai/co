package queue

import (
	"sync/atomic"
	"unsafe"
)

// Queue is a lock-free unbounded queue.
type MultiReceiverQueue[K any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer

	size     int32
	receiver uint32
}

type node[K any] struct {
	value K
	next  unsafe.Pointer

	dequeued uint32
}

type QueueReceiver[K any] struct {
	*MultiReceiverQueue[K]

	node unsafe.Pointer
}

// NewMultiReceiverQueue returns an empty queue.
func NewMultiReceiverQueue[K any]() *MultiReceiverQueue[K] {
	n := unsafe.Pointer(&node[K]{})
	return &MultiReceiverQueue[K]{head: n, tail: n, receiver: 1}
}

func (q *MultiReceiverQueue[K]) Receiver() *QueueReceiver[K] {
	atomic.AddUint32(&q.receiver, 1)
	return &QueueReceiver[K]{MultiReceiverQueue: q, node: nil}
}

func (q *MultiReceiverQueue[K]) len() int {
	return int(q.size)
}

// Enqueue puts the given value v at the tail of the queue.
func (q *MultiReceiverQueue[K]) enqueue(v K) {
	n := &node[K]{value: v}
	for {
		tail := load[K](&q.tail)
		next := load[K](&tail.next)
		if tail == load[K](&q.tail) { // are tail and next consistent?
			if next == nil {
				if cas(&tail.next, next, n) {
					atomic.AddInt32(&q.size, 1)
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
func (q *MultiReceiverQueue[K]) dequeue(r *QueueReceiver[K]) K {
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
				if r.node == nil {
					r.node = q.head
				}
				un := load[K](&r.node)
				unext := load[K](&un.next)
				if unext == nil {
					return *new(K)
				}
				cas(&r.node, un, unext)

				v := unext.value
				if atomic.AddUint32(&un.dequeued, 1) != q.receiver {
					atomic.AddInt32(&q.size, -1)
					return v
				}
				if cas(&q.head, head, next) {
					atomic.AddInt32(&q.size, -1)
					return v // Dequeue is done.  return
				}
			}
		}
	}
}

func (r *QueueReceiver[K]) Enqueue(v K) {
	r.MultiReceiverQueue.enqueue(v)
}

func (r *QueueReceiver[K]) Dequeue() K {
	return r.MultiReceiverQueue.dequeue(r)
}

func load[K any](p *unsafe.Pointer) (n *node[K]) {
	return (*node[K])(atomic.LoadPointer(p))
}

func cas[K any](p *unsafe.Pointer, old, new *node[K]) (ok bool) {
	return atomic.CompareAndSwapPointer(
		p, unsafe.Pointer(old), unsafe.Pointer(new))
}
