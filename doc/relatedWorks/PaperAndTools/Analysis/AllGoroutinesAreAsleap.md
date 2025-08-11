# All goroutines are asleap

Link: [Code](../../goPatch/src/runtime/proc.go#L6039)

## Summary

Go’s runtime scheduler continuously monitors the state of all goroutines. If, at any point, there are no runnable goroutines, the Go runtime detects a deadlock and panics.
To be precise, the runtime scheduler checks how many hardware threads $M$ are running. If the number is $0$ the runtime panics with the following message: "all goroutines are asleep - deadlock!".
This is only able to detect deadlocks, if it involves all go routines, a partial deadlock cannot be detected. If we assume in the following code snippet, that te locks in routine 0 and 2 run in a deadlock, it will not be detected by this mechanism, since the routine 1 continues to run.

```go
m := sync.Mutex{}
n := sync.Mutex{}

go func() {  // routine 1
  for {}
}

go func() {  // routine 2
  m.Lock()
  n.Lock()
}

// routine 0
n.Lock()
m.Lock()

```
