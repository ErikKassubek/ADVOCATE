# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixed4
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:91
```go
80 ...
81 
82 
83 	// Concurrent receives on unbuffered channel
84 	u := make(chan int)
85 	go func() { u <- 2 }()
86 	go func() { _ = <-u }() // A07 concurrently
87 	_ = <-u
88 
89 	// Possible recv on closed
90 	pc := make(chan int)
91 	close(pc)           // <-------
92 	_ = <-pc // P02
93 }
94 
```


###  Channel: Receive
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:92
```go
81 ...
82 
83 	// Concurrent receives on unbuffered channel
84 	u := make(chan int)
85 	go func() { u <- 2 }()
86 	go func() { _ = <-u }() // A07 concurrently
87 	_ = <-u
88 
89 	// Possible recv on closed
90 	pc := make(chan int)
91 	close(pc)
92 	_ = <-pc // P02           // <-------
93 }
94 
```


## Replay
**Replaying was not run**.

