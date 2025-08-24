# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCycleTwoLocks
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:16
```go
5 ...
6 
7 )
8 
9 // 1) Two‐lock inversion (classic two‐goroutine cycle)
10 func TestCycleTwoLocks(t *testing.T) {
11 	var a, b sync.Mutex
12 
13 	go func() {
14 		a.Lock()
15 		time.Sleep(10 * time.Millisecond)
16 		b.Lock()           // <-------
17 		b.Unlock()
18 		a.Unlock()
19 	}()
20 
21 	b.Lock()
22 	time.Sleep(5 * time.Millisecond)
23 	a.Lock()
24 	a.Unlock()
25 	b.Unlock()
26 }
27 
28 ...
```


## Replay
**Replaying was not run**.

