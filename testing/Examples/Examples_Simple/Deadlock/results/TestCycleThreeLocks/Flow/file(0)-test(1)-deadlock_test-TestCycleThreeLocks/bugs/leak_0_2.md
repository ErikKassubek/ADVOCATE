# Leak: Leak on sync.Mutex

The analyzer detected a leak on a sync.Mutex.
A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.
A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCycleThreeLocks
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Mutex: Lock
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


###  Mutex: Lock
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:40
```go
29 ...
30 
31 
32 	go func() {
33 		x.Lock()
34 		time.Sleep(5 * time.Millisecond)
35 		y.Lock()
36 		y.Unlock()
37 		x.Unlock()
38 	}()
39 	go func() {
40 		y.Lock()           // <-------
41 		time.Sleep(5 * time.Millisecond)
42 		z.Lock()
43 		z.Unlock()
44 		y.Unlock()
45 	}()
46 	// main goroutine forms the third link:
47 	z.Lock()
48 	time.Sleep(5 * time.Millisecond)
49 	x.Lock() // deadlock on x
50 	x.Unlock()
51 
52 ...
```


## Replay
The analyzer found a leak in the recorded trace.
The analyzer found a way to resolve the leak, meaning the leak should not reappear in the rewritten trace.

**Replaying was successful**.

It exited with the following code: 22

The replay was able to get the leaking mutex unstuck.

