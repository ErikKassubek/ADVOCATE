# Bug: Actual negative Wait Group

During the execution, a negative waitgroup counter occured.
The occurrence of a negative wait group counter lead to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixed2
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Waitgroup: Done
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:30
```go
19 ...
20 
21 	ch2 := make(chan int)
22 	close(ch2)
23 	ch2 <- 5 // A01
24 }
25 
26 // TestMixed2: A05 (negative wait group), A06 (unlock unlocked mutex) & L09 (waitgroup leak)
27 func TestMixed2(t *testing.T) {
28 	var wg sync.WaitGroup
29 	// Negative waitgroup counter
30 	wg.Done() // A05           // <-------
31 
32 	// Missing lock before unlock
33 	var mu sync.Mutex
34 	mu.Unlock() // A06
35 
36 	// Leak on waitgroup (never Done for this Add)
37 	wg.Add(1)
38 	go func() {
39 		// busy work
40 		time.Sleep(20 * time.Millisecond)
41 
42 ...
```


## Replay
**Replaying was not run**.

