# Bug: Actual close on nil channel

During the execution of the program, a close on a nil channel occurred.
The occurrence of a close on a nil channel lead to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA04_Mixed
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go:43
```go
32 ...
33 
34 	close(ch) // A03: may panic if closes cross
35 }
36 
37 func TestA04_Mixed(t *testing.T) {
38 	var ch chan int
39 	go func() {
40 		time.Sleep(10 * time.Millisecond)
41 		ch = make(chan int) // init too late
42 	}()
43 	close(ch) // A04: panic if nil when closed           // <-------
44 }
45 
```


## Replay
**Replaying was not run**.

