# Bug: Possible cyclic deadlock

The analysis detected a possible cyclic deadlock.
If this deadlock contains or influences the run of the main routine, this can result in the program getting stuck. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCycleTwoLocks
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Part of deadlock
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


###  Mutex: Causing deadlock
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
The bug is a potential bug.
The analyzer has tries to rewrite the trace in such a way, that the bug will be triggered when replaying the trace.

**Replaying failed**.

It exited with the following code: 10

The replay got stuck during the execution.
The main routine has already finished, but the trace still contains not executed operations.
This can be caused by a stuck replay.
Possible causes are:
    - The program was altered between recording and replay
    - The program execution path is not deterministic, e.g. its execution path is determined by a random number
    - The program execution path depends on the order of not tracked operations
    - The program execution depends on outside input, that was not exactly reproduced
	 - The program encountered a deadlock earlier in the trace than expected

