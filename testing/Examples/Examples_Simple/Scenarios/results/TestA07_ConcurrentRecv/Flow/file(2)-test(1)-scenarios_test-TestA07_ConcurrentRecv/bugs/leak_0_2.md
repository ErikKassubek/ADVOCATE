# Leak: Leak on unbuffered channel without possible partner

The analyzer detected a Leak on an unbuffered channel without a possible partner.
A Leak on an unbuffered channel is a situation, where a unbuffered channel is still blocking at the end of the program.
The analyzer could not find a partner for the stuck operation, which would resolve the leak.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA07_ConcurrentRecv
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:52
```go
41 ...
42 
43 }
44 
45 func TestA06_UnlockUnlocked(t *testing.T) {
46 	var mu sync.Mutex
47 	mu.Unlock() // A06
48 }
49 
50 func TestA07_ConcurrentRecv(t *testing.T) {
51 	ch := make(chan int)
52 	go func() { _ = <-ch }()           // <-------
53 	_ = <-ch // A07 (concurrent recv)
54 	close(ch)
55 }
56 
57 // ---------------- P01–P03: Possible Bugs ----------------
58 
59 func TestP01_PossibleSendOnClosed(t *testing.T) {
60 	ch := make(chan int)
61 	go func() {
62 		time.Sleep(10 * time.Millisecond)
63 
64 ...
```


## Replay
**Replaying was not run**.

