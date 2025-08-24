# Bug: Actual Send on Closed Channel

During the execution of the program, a send on a closed channel occurred.
The occurrence of a send on closed leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA01_SendOnClosed
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:18
```go
7 ...
8 
9 // ---------------- A00–A07: Absolute Bugs ----------------
10 
11 func TestA00_UnknownPanic(t *testing.T) {
12 	panic("triggering unknown panic") // A00
13 }
14 
15 func TestA01_SendOnClosed(t *testing.T) {
16 	ch := make(chan int)
17 	close(ch)
18 	ch <- 1 // A01           // <-------
19 }
20 
21 func TestA02_ReceiveOnClosed(t *testing.T) {
22 	ch := make(chan int)
23 	close(ch)
24 	_ = <-ch // A02 (Warning)
25 }
26 
27 func TestA03_CloseOnClosed(t *testing.T) {
28 	ch := make(chan int)
29 
30 ...
```


###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:17
```go
6 ...
7 
8 
9 // ---------------- A00–A07: Absolute Bugs ----------------
10 
11 func TestA00_UnknownPanic(t *testing.T) {
12 	panic("triggering unknown panic") // A00
13 }
14 
15 func TestA01_SendOnClosed(t *testing.T) {
16 	ch := make(chan int)
17 	close(ch)           // <-------
18 	ch <- 1 // A01
19 }
20 
21 func TestA02_ReceiveOnClosed(t *testing.T) {
22 	ch := make(chan int)
23 	close(ch)
24 	_ = <-ch // A02 (Warning)
25 }
26 
27 func TestA03_CloseOnClosed(t *testing.T) {
28 
29 ...
```


## Replay
**Replaying was not run**.

