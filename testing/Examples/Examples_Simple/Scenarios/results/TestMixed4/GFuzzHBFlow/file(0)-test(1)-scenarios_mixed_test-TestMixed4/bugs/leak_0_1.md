# Leak: Leak on unbuffered channel with possible partner

The analyzer detected a Leak on an unbuffered channel with a possible partner.
A Leak on an unbuffered channel is a situation, where a unbuffered channel is still blocking at the end of the program.
The partner is a corresponding send or receive operation, which communicated with another operation, but could communicated with the stuck operation instead, resolving the deadlock.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixed4
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:86
```go
75 ...
76 
77 	ch := make(chan int, 1)
78 	ch <- 1 // buffered
79 	go func() {
80 		_ = <-ch // partner exists (L03)
81 	}()
82 
83 	// Concurrent receives on unbuffered channel
84 	u := make(chan int)
85 	go func() { u <- 2 }()
86 	go func() { _ = <-u }() // A07 concurrently           // <-------
87 	_ = <-u
88 
89 	// Possible recv on closed
90 	pc := make(chan int)
91 	close(pc)
92 	_ = <-pc // P02
93 }
94 
```


###  Channel: Send
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:85
```go
74 ...
75 
76 	// Buffered channel with partner
77 	ch := make(chan int, 1)
78 	ch <- 1 // buffered
79 	go func() {
80 		_ = <-ch // partner exists (L03)
81 	}()
82 
83 	// Concurrent receives on unbuffered channel
84 	u := make(chan int)
85 	go func() { u <- 2 }()           // <-------
86 	go func() { _ = <-u }() // A07 concurrently
87 	_ = <-u
88 
89 	// Possible recv on closed
90 	pc := make(chan int)
91 	close(pc)
92 	_ = <-pc // P02
93 }
94 
```


## Replay
The analyzer found a leak in the recorded trace.
The analyzer found a way to resolve the leak, meaning the leak should not reappear in the rewritten trace.

**Replaying ended without confirming the bug**.

It exited with the following code: 0

The replay finished without being able to confirm the predicted bug. If the given trace was a directly recorded trace, this is the expected behavior. If it was rewritten by the analyzer, this could be an indication that something went wrong during rewrite or replay.

