# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA02_ReceiveOnClosed
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:23
```go
12 ...
13 
14 
15 func TestA01_SendOnClosed(t *testing.T) {
16 	ch := make(chan int)
17 	close(ch)
18 	ch <- 1 // A01
19 }
20 
21 func TestA02_ReceiveOnClosed(t *testing.T) {
22 	ch := make(chan int)
23 	close(ch)           // <-------
24 	_ = <-ch // A02 (Warning)
25 }
26 
27 func TestA03_CloseOnClosed(t *testing.T) {
28 	ch := make(chan int)
29 	close(ch)
30 	close(ch) // A03
31 }
32 
33 func TestA04_CloseOnNil(t *testing.T) {
34 
35 ...
```


###  Channel: Receive
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:24
```go
13 ...
14 
15 func TestA01_SendOnClosed(t *testing.T) {
16 	ch := make(chan int)
17 	close(ch)
18 	ch <- 1 // A01
19 }
20 
21 func TestA02_ReceiveOnClosed(t *testing.T) {
22 	ch := make(chan int)
23 	close(ch)
24 	_ = <-ch // A02 (Warning)           // <-------
25 }
26 
27 func TestA03_CloseOnClosed(t *testing.T) {
28 	ch := make(chan int)
29 	close(ch)
30 	close(ch) // A03
31 }
32 
33 func TestA04_CloseOnNil(t *testing.T) {
34 	var ch chan int
35 
36 ...
```


## Replay
**Replaying was not run**.

