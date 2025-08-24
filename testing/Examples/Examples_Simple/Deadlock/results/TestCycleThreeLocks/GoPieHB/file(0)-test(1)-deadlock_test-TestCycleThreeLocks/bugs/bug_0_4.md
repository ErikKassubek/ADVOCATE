# Bug: Possible cyclic deadlock

The analysis detected a possible cyclic deadlock.
If this deadlock contains or influences the run of the main routine, this can result in the program getting stuck. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCycleThreeLocks
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:42
```go
31 ...
32 
33 		x.Lock()
34 		time.Sleep(5 * time.Millisecond)
35 		y.Lock()
36 		y.Unlock()
37 		x.Unlock()
38 	}()
39 	go func() {
40 		y.Lock()
41 		time.Sleep(5 * time.Millisecond)
42 		z.Lock()           // <-------
43 		z.Unlock()
44 		y.Unlock()
45 	}()
46 	// main goroutine forms the third link:
47 	z.Lock()
48 	time.Sleep(5 * time.Millisecond)
49 	x.Lock() // deadlock on x
50 	x.Unlock()
51 	z.Unlock()
52 }
53 
54 ...
```


###  Mutex: Part of deadlock
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:49
```go
38 ...
39 
40 		y.Lock()
41 		time.Sleep(5 * time.Millisecond)
42 		z.Lock()
43 		z.Unlock()
44 		y.Unlock()
45 	}()
46 	// main goroutine forms the third link:
47 	z.Lock()
48 	time.Sleep(5 * time.Millisecond)
49 	x.Lock() // deadlock on x           // <-------
50 	x.Unlock()
51 	z.Unlock()
52 }
53 
54 // 3) Channel‐based deadlock: send+recv missing
55 func TestChannelDeadlock(t *testing.T) {
56 	ch := make(chan int)
57 	// no corresponding receive
58 	ch <- 1
59 }
60 
61 ...
```


-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:35
```go
24 ...
25 
26 }
27 
28 // 2) Three‐lock cycle (A→B→C→A)
29 func TestCycleThreeLocks(t *testing.T) {
30 	var x, y, z sync.Mutex
31 
32 	go func() {
33 		x.Lock()
34 		time.Sleep(5 * time.Millisecond)
35 		y.Lock()           // <-------
36 		y.Unlock()
37 		x.Unlock()
38 	}()
39 	go func() {
40 		y.Lock()
41 		time.Sleep(5 * time.Millisecond)
42 		z.Lock()
43 		z.Unlock()
44 		y.Unlock()
45 	}()
46 
47 ...
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

