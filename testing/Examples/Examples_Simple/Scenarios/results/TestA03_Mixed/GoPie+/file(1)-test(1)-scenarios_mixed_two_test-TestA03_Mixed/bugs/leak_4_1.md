# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA03_Mixed
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go:29
```go
18 ...
19 
20 	ch := make(chan int)
21 	go func() {
22 		time.Sleep(10 * time.Millisecond)
23 		close(ch) // close before receive
24 	}()
25 	_ = <-ch // A02: might receive from closed
26 }
27 
28 func TestA03_Mixed(t *testing.T) {
29 	ch := make(chan int)           // <-------
30 	go func() {
31 		time.Sleep(10 * time.Millisecond)
32 		close(ch)
33 	}()
34 	close(ch) // A03: may panic if closes cross
35 }
36 
37 func TestA04_Mixed(t *testing.T) {
38 	var ch chan int
39 	go func() {
40 
41 ...
```


## Replay
**Replaying was not run**.

