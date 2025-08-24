# Leak: Leak on sync.WaitGroup

The analyzer detected a leak on a sync.WaitGroup.
A leak on a sync.WaitGroup is a situation, where a sync.WaitGroup is still blocking at the end of the program.
A sync.WaitGroup wait is blocking, because the counter is not zero.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestWaitGroupDeadlock
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Waitgroup: Wait
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

