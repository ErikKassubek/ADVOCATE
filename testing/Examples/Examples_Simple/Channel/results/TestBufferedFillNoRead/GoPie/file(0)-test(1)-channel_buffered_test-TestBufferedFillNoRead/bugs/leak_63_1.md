# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestBufferedFillNoRead
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Channel/channel_buffered_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Channel/channel_buffered_test.go:19
```go
8 ...
9 
10 // ================================
11 
12 // TestBufferedFillNoRead: buffered channel full, no reader
13 func TestBufferedFillNoRead(t *testing.T) {
14 	ch := make(chan int, 2)
15 	ch <- 1
16 	ch <- 2 // now full
17 	go func() {
18 		// no reader
19 		ch <- 3 // L04 (buffered leak no partner)           // <-------
20 	}()
21 	time.Sleep(10 * time.Millisecond)
22 }
23 
24 // TestBufferedDrainSlow: buffered channel drained slowly
25 func TestBufferedDrainSlow(t *testing.T) {
26 	ch := make(chan int, 2)
27 	ch <- 1
28 	ch <- 2
29 	go func() {
30 
31 ...
```


## Replay
**Replaying was not run**.

