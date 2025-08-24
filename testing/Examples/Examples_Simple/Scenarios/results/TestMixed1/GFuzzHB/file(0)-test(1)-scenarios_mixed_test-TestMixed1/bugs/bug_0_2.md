# Bug: Actual Send on Closed Channel

During the execution of the program, a send on a closed channel occurred.
The occurrence of a send on closed leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixed1
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:23
```go
12 ...
13 
14 	ch1 := make(chan int)
15 	go func() {
16 		// no receiver for ch1, this goroutine will leak
17 		ch1 <- 1 // L02
18 	}()
19 
20 	// Closed channel panic
21 	ch2 := make(chan int)
22 	close(ch2)
23 	ch2 <- 5 // A01           // <-------
24 }
25 
26 // TestMixed2: A05 (negative wait group), A06 (unlock unlocked mutex) & L09 (waitgroup leak)
27 func TestMixed2(t *testing.T) {
28 	var wg sync.WaitGroup
29 	// Negative waitgroup counter
30 	wg.Done() // A05
31 
32 	// Missing lock before unlock
33 	var mu sync.Mutex
34 
35 ...
```


###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:22
```go
11 ...
12 
13 	// Unbuffered channel leak
14 	ch1 := make(chan int)
15 	go func() {
16 		// no receiver for ch1, this goroutine will leak
17 		ch1 <- 1 // L02
18 	}()
19 
20 	// Closed channel panic
21 	ch2 := make(chan int)
22 	close(ch2)           // <-------
23 	ch2 <- 5 // A01
24 }
25 
26 // TestMixed2: A05 (negative wait group), A06 (unlock unlocked mutex) & L09 (waitgroup leak)
27 func TestMixed2(t *testing.T) {
28 	var wg sync.WaitGroup
29 	// Negative waitgroup counter
30 	wg.Done() // A05
31 
32 	// Missing lock before unlock
33 
34 ...
```


## Replay
**Replaying was not run**.

