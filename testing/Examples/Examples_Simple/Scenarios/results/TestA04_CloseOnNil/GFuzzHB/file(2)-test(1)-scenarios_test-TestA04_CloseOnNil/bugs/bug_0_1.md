# Bug: Actual close on nil channel

During the execution of the program, a close on a nil channel occurred.
The occurrence of a close on a nil channel lead to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA04_CloseOnNil
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:35
```go
24 ...
25 
26 
27 func TestA03_CloseOnClosed(t *testing.T) {
28 	ch := make(chan int)
29 	close(ch)
30 	close(ch) // A03
31 }
32 
33 func TestA04_CloseOnNil(t *testing.T) {
34 	var ch chan int
35 	close(ch) // A04           // <-------
36 }
37 
38 func TestA05_NegativeWaitGroup(t *testing.T) {
39 	var wg sync.WaitGroup
40 	wg.Add(1)
41 	wg.Done()
42 	wg.Done() // A05
43 }
44 
45 func TestA06_UnlockUnlocked(t *testing.T) {
46 
47 ...
```


## Replay
**Replaying was not run**.

