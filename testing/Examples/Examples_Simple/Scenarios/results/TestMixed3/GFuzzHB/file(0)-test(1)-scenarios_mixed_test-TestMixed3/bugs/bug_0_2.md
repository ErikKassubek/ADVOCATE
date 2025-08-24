# Bug: Actual Send on Closed Channel

During the execution of the program, a send on a closed channel occurred.
The occurrence of a send on closed leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixed3
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:61
```go
50 ...
51 
52 		wg.Done()
53 		wg.Done() // P03
54 	}()
55 
56 	// Possible send on closed
57 	ch := make(chan int)
58 	go func() {
59 		close(ch)
60 	}()
61 	ch <- 10 // P01           // <-------
62 
63 	// Leak on mutex: lock twice without unlock
64 	var mu sync.Mutex
65 	mu.Lock()
66 	go func() {
67 		// no unlock for outer lock
68 		mu.Unlock()
69 	}()
70 	mu.Lock() // L08 if scheduling orders leak
71 	mu.Unlock()
72 
73 ...
```


###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:59
```go
48 ...
49 
50 	// Possible negative waitgroup inside goroutine
51 	go func() {
52 		wg.Done()
53 		wg.Done() // P03
54 	}()
55 
56 	// Possible send on closed
57 	ch := make(chan int)
58 	go func() {
59 		close(ch)           // <-------
60 	}()
61 	ch <- 10 // P01
62 
63 	// Leak on mutex: lock twice without unlock
64 	var mu sync.Mutex
65 	mu.Lock()
66 	go func() {
67 		// no unlock for outer lock
68 		mu.Unlock()
69 	}()
70 
71 ...
```


## Replay
**Replaying was not run**.

