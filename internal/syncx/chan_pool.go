package syncx

import "sync"

var chanPool = sync.Pool{New: func() interface{} { return make(chan struct{}) }}
