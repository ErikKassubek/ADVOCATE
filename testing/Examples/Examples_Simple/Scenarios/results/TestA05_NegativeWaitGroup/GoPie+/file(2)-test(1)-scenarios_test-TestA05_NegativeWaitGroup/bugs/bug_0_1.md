# Bug: Actual negative Wait Group

During the execution, a negative waitgroup counter occured.
The occurrence of a negative wait group counter lead to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA05_NegativeWaitGroup
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Waitgroup: Done
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:42
```go
31 ...
32 
33 func TestA04_CloseOnNil(t *testing.T) {
34 	var ch chan int
35 	close(ch) // A04
36 }
37 
38 func TestA05_NegativeWaitGroup(t *testing.T) {
39 	var wg sync.WaitGroup
40 	wg.Add(1)
41 	wg.Done()
42 	wg.Done() // A05           // <-------
43 }
44 
45 func TestA06_UnlockUnlocked(t *testing.T) {
46 	var mu sync.Mutex
47 	mu.Unlock() // A06
48 }
49 
50 func TestA07_ConcurrentRecv(t *testing.T) {
51 	ch := make(chan int)
52 	go func() { _ = <-ch }()
53 
54 ...
```


## Replay
**Replaying was not run**.

