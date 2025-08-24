# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA02_Mixed
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go:23
```go
12 ...
13 
14 		close(ch) // closes while another routine sends
15 	}()
16 	ch <- 1 // A01: may send on closed
17 }
18 
19 func TestA02_Mixed(t *testing.T) {
20 	ch := make(chan int)
21 	go func() {
22 		time.Sleep(10 * time.Millisecond)
23 		close(ch) // close before receive           // <-------
24 	}()
25 	_ = <-ch // A02: might receive from closed
26 }
27 
28 func TestA03_Mixed(t *testing.T) {
29 	ch := make(chan int)
30 	go func() {
31 		time.Sleep(10 * time.Millisecond)
32 		close(ch)
33 	}()
34 
35 ...
```


###  Channel: Receive
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_two_test.go:25
```go
14 ...
15 
16 	ch <- 1 // A01: may send on closed
17 }
18 
19 func TestA02_Mixed(t *testing.T) {
20 	ch := make(chan int)
21 	go func() {
22 		time.Sleep(10 * time.Millisecond)
23 		close(ch) // close before receive
24 	}()
25 	_ = <-ch // A02: might receive from closed           // <-------
26 }
27 
28 func TestA03_Mixed(t *testing.T) {
29 	ch := make(chan int)
30 	go func() {
31 		time.Sleep(10 * time.Millisecond)
32 		close(ch)
33 	}()
34 	close(ch) // A03: may panic if closes cross
35 }
36 
37 ...
```


## Replay
**Replaying was not run**.

