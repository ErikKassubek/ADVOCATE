# Bug: Actual Close on Closed Channel

During the execution of the program, a close on a close channel occurred.
The occurrence of a close on a closed channel lead to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA03_Mixed
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go:32
```go
21 ...
22 
23 		close(ch) // close before receive
24 	}()
25 	_ = <-ch // A02: might receive from closed
26 }
27 
28 func TestA03_Mixed(t *testing.T) {
29 	ch := make(chan int)
30 	go func() {
31 		time.Sleep(10 * time.Millisecond)
32 		close(ch)           // <-------
33 	}()
34 	close(ch) // A03: may panic if closes cross
35 }
36 
37 func TestA04_Mixed(t *testing.T) {
38 	var ch chan int
39 	go func() {
40 		time.Sleep(10 * time.Millisecond)
41 		ch = make(chan int) // init too late
42 	}()
43 
44 ...
```


###  Channel: Close
## Replay
**Replaying was not run**.

