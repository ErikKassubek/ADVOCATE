# Bug: Actual unlock of not locked mutex

During the execution, a not locked mutex was unlocked.
The occurrence of this lead to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA06_UnlockUnlocked
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Lock
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:47
```go
36 ...
37 
38 func TestA05_NegativeWaitGroup(t *testing.T) {
39 	var wg sync.WaitGroup
40 	wg.Add(1)
41 	wg.Done()
42 	wg.Done() // A05
43 }
44 
45 func TestA06_UnlockUnlocked(t *testing.T) {
46 	var mu sync.Mutex
47 	mu.Unlock() // A06           // <-------
48 }
49 
50 func TestA07_ConcurrentRecv(t *testing.T) {
51 	ch := make(chan int)
52 	go func() { _ = <-ch }()
53 	_ = <-ch // A07 (concurrent recv)
54 	close(ch)
55 }
56 
57 // ---------------- P01–P03: Possible Bugs ----------------
58 
59 ...
```


## Replay
**Replaying was not run**.

