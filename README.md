# Co

A concurrency related (sort of) project dedicate to help developer ease the pain of dealing goroutine and 
channel when coding 2 more lines channel code become annoying.

`Co` provides a lot of (at least planning) helper function's and utils such as `AwaitAll` `AwaitAny` and 
Parallel execution mechanism with concurrency limitation.

If there are some concurrency patterns you would like me to implement and there is none available in
community or in go lib, you are more than welcome to open an issue.

## Listence

MIT

## Doc

https://godoc.org/github.com/tempura-shrimp/co

Note: there are a lot interface{} in project and types, which i do not like, but since Go have not yet
implemented Generic, this is what I need to do in order to make response work. Once Go release a version
with Generic (some day), I'll adapt that ASAP.

## Examples

### Parallel

```golang
// replace with `co.NewParallel` if no response needed
p := co.NewParallelWithResponse(10) // worker size
for i := 0; i < 10000; i++ {
    i := i
    p.AddWithResponse(func() interface{} {
        return i + 1
    })
}

vals := p.Wait()
```

### Awaits

```golang

handlers := make([]func() interface{}, 0)
for i := 0; i < 1000; i++ {
    i := i
    handlers = append(handlers, func() interface{} {
        return i + 1
    })
}

responses := co.AwaitAll(handlers...)
```