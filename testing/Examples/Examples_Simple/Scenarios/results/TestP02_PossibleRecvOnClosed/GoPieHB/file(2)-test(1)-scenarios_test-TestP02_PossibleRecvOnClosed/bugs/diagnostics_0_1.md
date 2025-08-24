# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestP02_PossibleRecvOnClosed
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:76
```go
65 ...
66 
67 
68 func TestP02_PossibleRecvOnClosed(t *testing.T) {
69 	ch := make(chan int)
70 	go func() {
71 		ch <- 1
72 		close(ch)
73 	}()
74 	time.Sleep(10 * time.Millisecond)
75 	_ = <-ch // P02
76 	_ = <-ch // P02           // <-------
77 }
78 
79 func TestP03_PossibleNegativeWaitGroup(t *testing.T) {
80 	var wg sync.WaitGroup
81 	wg.Add(1)
82 	go func() {
83 		wg.Done()
84 		wg.Done() // P03
85 	}()
86 	wg.Wait()
87 
88 ...
```


###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:72
```go
61 ...
62 
63 		close(ch)
64 	}()
65 	ch <- 42 // P01 (race with close)
66 }
67 
68 func TestP02_PossibleRecvOnClosed(t *testing.T) {
69 	ch := make(chan int)
70 	go func() {
71 		ch <- 1
72 		close(ch)           // <-------
73 	}()
74 	time.Sleep(10 * time.Millisecond)
75 	_ = <-ch // P02
76 	_ = <-ch // P02
77 }
78 
79 func TestP03_PossibleNegativeWaitGroup(t *testing.T) {
80 	var wg sync.WaitGroup
81 	wg.Add(1)
82 	go func() {
83 
84 ...
```


## Replay
**Replaying was not run**.

