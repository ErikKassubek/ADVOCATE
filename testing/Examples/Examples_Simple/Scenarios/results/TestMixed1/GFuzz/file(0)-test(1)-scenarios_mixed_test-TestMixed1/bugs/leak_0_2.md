# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixed1
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:17
```go
6 ...
7 
8 
9 // ---------------- Mixed scenarios ----------------
10 
11 // TestMixed1: A01 (send on closed channel) & L02 (unbuffered leak without partner)
12 func TestMixed1(t *testing.T) {
13 	// Unbuffered channel leak
14 	ch1 := make(chan int)
15 	go func() {
16 		// no receiver for ch1, this goroutine will leak
17 		ch1 <- 1 // L02           // <-------
18 	}()
19 
20 	// Closed channel panic
21 	ch2 := make(chan int)
22 	close(ch2)
23 	ch2 <- 5 // A01
24 }
25 
26 // TestMixed2: A05 (negative wait group), A06 (unlock unlocked mutex) & L09 (waitgroup leak)
27 func TestMixed2(t *testing.T) {
28 
29 ...
```


## Replay
**Replaying was not run**.

