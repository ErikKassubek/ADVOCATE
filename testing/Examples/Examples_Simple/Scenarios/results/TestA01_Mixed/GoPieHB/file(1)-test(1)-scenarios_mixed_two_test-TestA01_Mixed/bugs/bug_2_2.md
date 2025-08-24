# Bug: Actual Send on Closed Channel

During the execution of the program, a send on a closed channel occurred.
The occurrence of a send on closed leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA01_Mixed
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go:16
```go
5 ...
6 
7 
8 // ----------
9 
10 func TestA01_Mixed(t *testing.T) {
11 	ch := make(chan int)
12 	go func() {
13 		time.Sleep(10 * time.Millisecond)
14 		close(ch) // closes while another routine sends
15 	}()
16 	ch <- 1 // A01: may send on closed           // <-------
17 }
18 
19 func TestA02_Mixed(t *testing.T) {
20 	ch := make(chan int)
21 	go func() {
22 		time.Sleep(10 * time.Millisecond)
23 		close(ch) // close before receive
24 	}()
25 	_ = <-ch // A02: might receive from closed
26 }
27 
28 ...
```


###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go:14
```go
3 ...
4 
5 	"time"
6 )
7 
8 // ----------
9 
10 func TestA01_Mixed(t *testing.T) {
11 	ch := make(chan int)
12 	go func() {
13 		time.Sleep(10 * time.Millisecond)
14 		close(ch) // closes while another routine sends           // <-------
15 	}()
16 	ch <- 1 // A01: may send on closed
17 }
18 
19 func TestA02_Mixed(t *testing.T) {
20 	ch := make(chan int)
21 	go func() {
22 		time.Sleep(10 * time.Millisecond)
23 		close(ch) // close before receive
24 	}()
25 
26 ...
```


## Replay
**Replaying was not run**.

