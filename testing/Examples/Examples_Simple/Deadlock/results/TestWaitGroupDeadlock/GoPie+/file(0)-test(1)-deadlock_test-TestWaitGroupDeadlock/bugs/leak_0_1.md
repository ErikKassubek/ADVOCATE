# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestWaitGroupDeadlock
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:69
```go
58 ...
59 
60 
61 // 4) WaitGroup deadlock: missing Done
62 func TestWaitGroupDeadlock(t *testing.T) {
63 	var wg sync.WaitGroup
64 	wg.Add(1)
65 	go func() {
66 		// forgot wg.Done()
67 		time.Sleep(20 * time.Millisecond)
68 	}()
69 	wg.Wait() // blocks forever           // <-------
70 }
71 
72 // 5) Cond deadlock: Wait with no Signal
73 func TestCondDeadlock(t *testing.T) {
74 	var mu sync.Mutex
75 	cond := sync.NewCond(&mu)
76 	go func() {
77 		time.Sleep(10 * time.Millisecond)
78 		// no cond.Signal()
79 	}()
80 
81 ...
```


## Replay
**Replaying was not run**.

