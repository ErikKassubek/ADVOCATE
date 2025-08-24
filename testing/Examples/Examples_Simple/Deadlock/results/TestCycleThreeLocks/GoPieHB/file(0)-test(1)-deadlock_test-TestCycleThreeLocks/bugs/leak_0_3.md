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


###  Mutex: Lock
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:47
```go
36 ...
37 
38 	}()
39 	go func() {
40 		y.Lock()
41 		time.Sleep(5 * time.Millisecond)
42 		z.Lock()
43 		z.Unlock()
44 		y.Unlock()
45 	}()
46 	// main goroutine forms the third link:
47 	z.Lock()           // <-------
48 	time.Sleep(5 * time.Millisecond)
49 	x.Lock() // deadlock on x
50 	x.Unlock()
51 	z.Unlock()
52 }
53 
54 // 3) Channel‐based deadlock: send+recv missing
55 func TestChannelDeadlock(t *testing.T) {
56 	ch := make(chan int)
57 	// no corresponding receive
58 
59 ...
```


## Replay
The analyzer found a leak in the recorded trace.
The analyzer found a way to resolve the leak, meaning the leak should not reappear in the rewritten trace.

**Replaying was successful**.

It exited with the following code: 22

The replay was able to get the leaking mutex unstuck.

