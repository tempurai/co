[![Go](https://github.com/tempura-shrimp/co/actions/workflows/go.yml/badge.svg)](https://github.com/tempura-shrimp/co/actions/workflows/go.yml)

# Co

**Co** is a concurrency project with GENERIC SUPPORTED, dedicate to three things:

1. Providing mechanism to dealing data in ReactiveX fashion and related transform algorithm.

- Async sequence with transform functions such as `map`, `filter`, `multicast`, `buffer_time` and others.

2. Providing a round trip pipe on top of Async sequence.

3. Helping developer to ease the pain of dealing goroutine and channel with less than 2 lines code with:

- mimicked promising functions: `AwaitAll`, `AwaitAny`, `AwaitRace`
- high performance none blocking queue: `Queue`, `MultiReceiverQueue`
- high performance worker pool: `WorkerPool`, `DispatchePool`

## Motivation on Go with Reactive programming

I consider ReactiveX programming pattern as a data stream friendly way to dealing with never ending data. However, most common scenario for such a pattern is in client side programming, for example, in AngularX, and I actually never saw any backend project using the pattern. Actually, it's quite easy to understand. The common pattern in backend simply is to push something to controller (and it will send some data to database) and listen to the callback.

But, considering a case where server should continue to receive time series data, such as user log. The normal pattern would be having an API server to listen to the data, process it, and then send it to some message queue. Usually it's one to one pattern, namely each incoming user log mapped to one request to message queue. With the size of log data, the size of request message and combing with load to send data in TCP pipe, we could just save the user log to some size and send them all together. In this case we have an optimized user log processing pipeline.

I believe that ReactiveX is a good way to deal with such a case. Of course, increasing people are using or considering Go for client side programming, which I believe definitely should apply Co.

However, even though I have mentioned a lot of ReactiveX patterns above. I do not want to create something with exact API. It's due to 1. I found it's originally API to be hard to understand; 2. The server side programming usually don't require that much of time based algorithm.

## APIs

https://godoc.org/github.com/tempura-shrimp/co

### Promising functions:

- `AwaitAll`: wait for all promises to be resolved or failed.
- `AwaitRace`: wait for any promises to be resolved.
- `AwaitAny`: wait for any promises to be resolved or failed.

### Queue:

- `Queue`: a none blocking queue with unlimited size.
- `MultiReceiverQueue`: multiple receiver version of `Queue`.

### Async Sequence:

#### Combination

- `AsyncCombineLatest`: combine latest of multiple async sequence with different types.
- `AsyncMerged`: merge multiple async sequence with same type.
- `AsyncMultiCast`: broadcasting async sequence to multiple successor sequences.
- `AsyncPartition`: horizontally partition elements of multiple async sequence.
- `AsyncZip`: get the latest result of all multiple async sequence with different type.

#### Transform

- `AsyncAdjacentFilter`: filter adjacent elements.
- `AsyncBufferTime`: buffer elements for a certain time.
- `AsyncCompacted`: remove empty value form elements.
- `AsyncFlatten`: flatten nested async sequence.
- `AsyncMap`: map elements to other type / value.

#### Time based transform

- `AsyncDebounce`: discard elements inside or outside a given sliding windows.

#### Creating asynchronous sequence

- `OfList`: create an asynchronous sequence from a list.
- `FromChan`: create an asynchronous sequence from a channel; also can be created with a buffered channel.
- `FromChanBuffered`: create an asynchronous sequence from a channel with unlimited buffer size.
- `FromFn`: create an asynchronous sequence from closure function.
- `AsyncSubject`: create an asynchronous sequence with a Next/Error/Complete method.

#### Round Trip

- `AsyncRoundTripper`: create an asynchronous manager with a round trip, which mean sender can receive callback from handler, it can be used to create an HTTP server.

## Getting started

Navigate to your project base and `go get github.com/tempura-shrimp/co`

## Examples

### Parallel

```golang
p := co.NewParallel[bool](10)// worker size
for i := 0; i < 10000; i++ {
    func(idx int) {
        p.Process(func() bool {
            actual[idx] = true
           return true
        })
    }(i)
}

// Wait doesn't indicate a Run, the job will run once added
// convey.So, you could ignore Wait() in some cases
vals := p.Wait()
```

### Awaits

```golang
handlers := make([]func() (int, error), 0)
for i := 0; i < 1000; i++ {
    i := i
    handlers = append(handlers, func() (int, error) {
        return i + 1, nil
    })
}

responses := co.AwaitAll[int](handlers...)
```

### Async Sequence

```golang
numbers := []int{1, 4, 5, 6, 7}
aList := co.OfListWith(numbers...)

numbers2 := []int{2, 4, 7, 0, 21}
aList2 := co.OfListWith(numbers2...)
mList := co.NewAsyncMapSequence[int](aList, func(v int) int {
    return v + 1
})

pList := co.NewAsyncMergedSequence[int](mList, aList2)

result := []int{}
for data := range pList.Iter() {
    result = append(result, data)
}
```

with time based transformation

```golang
queued := []int{1, 4, 5, 6, 7, 2, 2, 3, 4, 5, 12, 4, 2, 3, 43, 127, 37598, 34, 34, 123, 123}
sourceCh := make(chan int)

oChannel := co.FromChan(sourceCh)
bList := co.NewAsyncBufferTimeSequence[int](oChannel, time.Second)

// simulate handling on other go routine
go func() {
    time.Sleep(time.Second)
    for i, val := range queued {
        sourceCh <- val
        time.Sleep(time.Millisecond * (100 + time.Duration(i)*10))
    }
    oChannel.Complete()
}()

result := [][]int{}
for data := range bList.Iter() {
    result = append(acturesultal, data)
}
```

## Benchmark

```
goos: darwin
goarch: amd64
pkg: github.com/tempura-shrimp/co/pool
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkSimpleWorkPoolWithFib-12          10000            176316 ns/op
BenchmarkFibSequence-12                    10000            959778 ns/op
BenchmarkFibWithAwaitAll-12                10000            194145 ns/op
BenchmarkFibWithTunny-12                   10000            907005 ns/op
BenchmarkHeavyWorkPoolWithFib-12           10000            178232 ns/op
PASS
```
