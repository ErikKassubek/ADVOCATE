# Leak: Leak on unbuffered channel without possible partner

The analyzer detected a Leak on an unbuffered channel without a possible partner.
A Leak on an unbuffered channel is a situation, where a unbuffered channel is still blocking at the end of the program.
The analyzer could not find a partner for the stuck operation, which would resolve the leak.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestBufferedFillNoRead
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Channel/channel_buffered_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Channel: Send
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

