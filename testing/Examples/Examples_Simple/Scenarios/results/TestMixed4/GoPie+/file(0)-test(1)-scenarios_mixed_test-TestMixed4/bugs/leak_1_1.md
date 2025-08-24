# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixed4
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
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


## Replay
**Replaying was not run**.

