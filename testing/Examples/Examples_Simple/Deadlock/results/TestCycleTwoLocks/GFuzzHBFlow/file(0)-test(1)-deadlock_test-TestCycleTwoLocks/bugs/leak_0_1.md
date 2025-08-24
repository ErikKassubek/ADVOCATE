# Leak: Leak on sync.Mutex

The analyzer detected a leak on a sync.Mutex.
A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.
A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCycleTwoLocks
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Mutex: Lock
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:23
```go
12 ...
13 
14 		a.Lock()
15 		time.Sleep(10 * time.Millisecond)
16 		b.Lock()
17 		b.Unlock()
18 		a.Unlock()
19 	}()
20 
21 	b.Lock()
22 	time.Sleep(5 * time.Millisecond)
23 	a.Lock()           // <-------
24 	a.Unlock()
25 	b.Unlock()
26 }
27 
28 // 2) Three‐lock cycle (A→B→C→A)
29 func TestCycleThreeLocks(t *testing.T) {
30 	var x, y, z sync.Mutex
31 
32 	go func() {
33 		x.Lock()
34 
35 ...
```


###  Mutex: Lock
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:14
```go
3 ...
4 
5 	"testing"
6 	"time"
7 )
8 
9 // 1) Two‐lock inversion (classic two‐goroutine cycle)
10 func TestCycleTwoLocks(t *testing.T) {
11 	var a, b sync.Mutex
12 
13 	go func() {
14 		a.Lock()           // <-------
15 		time.Sleep(10 * time.Millisecond)
16 		b.Lock()
17 		b.Unlock()
18 		a.Unlock()
19 	}()
20 
21 	b.Lock()
22 	time.Sleep(5 * time.Millisecond)
23 	a.Lock()
24 	a.Unlock()
25 
26 ...
```


## Replay
The analyzer found a leak in the recorded trace.
The analyzer found a way to resolve the leak, meaning the leak should not reappear in the rewritten trace.

**Replaying was successful**.

It exited with the following code: 22

The replay was able to get the leaking mutex unstuck.

