package queue

type Queue[K any] struct {
	r *QueueReceiver[K]
}

func NewQueue[K any]() *Queue[K] {
	return &Queue[K]{r: NewMultiReceiverQueue[K]().Receiver()}
}

func (q *Queue[K]) Len() int {
	return q.r.len()
}

func (q *Queue[K]) Enqueue(v K) {
	q.r.Enqueue(v)
}

func (q *Queue[K]) Dequeue() K {
	return q.r.Dequeue()
}
