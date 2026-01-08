package queue

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// Queue is a lock-free unbounded queue.
type MultiReceiverQueue[K any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer

	size     int32
	receiver uint32

	nodeCache sync.Pool
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
	r := &MultiReceiverQueue[K]{head: n, tail: n, receiver: 0}
	r.nodeCache.New = func() any { return &node[K]{} }
	return r
}

func (q *MultiReceiverQueue[K]) Receiver() *QueueReceiver[K] {
	atomic.AddUint32(&q.receiver, 1)
	return &QueueReceiver[K]{MultiReceiverQueue: q, node: q.head}
}

func (q *MultiReceiverQueue[K]) Count() int {
	return int(q.size)
}

// Enqueue puts the given value v at the tail of the queue.
func (q *MultiReceiverQueue[K]) Enqueue(v K) {
	n := q.nodeCache.Get().(*node[K])
	n.value = v
	atomic.StoreUint32(&n.dequeued, 0)
	atomic.StorePointer(&n.next, nil)

	for {
		tail := load[K](&q.tail)
		next := load[K](&tail.next)
		if tail == load[K](&q.tail) { // are tail and next consistent?
			if next == nil {
				if cas(&tail.next, next, n) {
					cas(&q.tail, tail, n) // Enqueue is done.  try to swing tail to the inserted node
					atomic.AddInt32(&q.size, 1)
					return
				}
			} else { // tail was not pointing to the last node
				// try to swing Tail to the next node
				cas(&q.tail, tail, next)
			}
		}
	}
}

func (q *MultiReceiverQueue[K]) purgeNode(n *node[K]) {
	atomic.AddInt32(&q.size, -1)
	atomic.StoreUint32(&n.dequeued, 0)
	atomic.StorePointer(&n.next, nil)
	q.nodeCache.Put(n) // head is dead, put back to pool
}

// Dequeue removes and returns the value at the head of the queue.
// It returns false if the queue is empty.
func (q *MultiReceiverQueue[K]) dequeue(r *QueueReceiver[K]) K {
	for {
		// advacing head
		head := load[K](&q.head)
		next := load[K](&head.next)

		tail := load[K](&q.tail)

		if head == load[K](&q.head) { // preflight check
			if head == tail { // is queue empty or tail falling behind?
				if next == nil { // is queue empty?
					return *new(K)
				}
				// tail is falling behind.  try to advance it
				cas(&q.tail, tail, next)

			} else if atomic.LoadUint32(&next.dequeued) == r.receiver {
				if next == nil { // is queue empty?
					return *new(K)
				}
				if cas(&q.head, head, next) {
					q.purgeNode(head) // head is dead, put back to pool
				}
			} else {
				rhead := load[K](&r.node)
				rnext := load[K](&rhead.next)

				if rhead == load[K](&r.node) {
					if rnext == nil { // is queue empty?
						return *new(K)
					}

					if cas(&r.node, rhead, rnext) {
						v := rnext.value // Read value after successful CAS
						if atomic.AddUint32(&rnext.dequeued, 1) == q.receiver {
							// where the receiver is the last receiver in the queue
							if cas(&q.head, rhead, rnext) {
								// if rhead is q.head, then try to advance q.head
								q.purgeNode(rhead) // head is dead, put back to pool
							}
						}
						return v // Dequeue is done.  return
					}
				}
			}
		}
	}
}

func (r *QueueReceiver[K]) Enqueue(v K) {
	r.MultiReceiverQueue.Enqueue(v)
}

func (r *QueueReceiver[K]) Dequeue() K {
	return r.MultiReceiverQueue.dequeue(r)
}

func (r *QueueReceiver[K]) IsEmpty() bool {
	return r.node == r.tail
}

func load[K any](p *unsafe.Pointer) (n *node[K]) {
	return (*node[K])(atomic.LoadPointer(p))
}

func cas[K any](p *unsafe.Pointer, old, new *node[K]) (ok bool) {
	return atomic.CompareAndSwapPointer(
		p, unsafe.Pointer(old), unsafe.Pointer(new))
}
