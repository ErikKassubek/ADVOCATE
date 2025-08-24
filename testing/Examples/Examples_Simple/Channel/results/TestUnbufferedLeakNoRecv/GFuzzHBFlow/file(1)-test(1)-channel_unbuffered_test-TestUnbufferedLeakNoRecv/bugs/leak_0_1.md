# Leak: Leak on unbuffered channel without possible partner

The analyzer detected a Leak on an unbuffered channel without a possible partner.
A Leak on an unbuffered channel is a situation, where a unbuffered channel is still blocking at the end of the program.
The analyzer could not find a partner for the stuck operation, which would resolve the leak.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestUnbufferedLeakNoRecv
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Channel/channel_unbuffered_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Channel: Send
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Channel/channel_unbuffered_test.go:24
```go
13 ...
14 
15 	go func() { ch <- 1 }()
16 	_ = <-ch
17 }
18 
19 // TestUnbufferedLeakNoRecv: send blocks forever
20 func TestUnbufferedLeakNoRecv(t *testing.T) {
21 	ch := make(chan int)
22 	go func() {
23 		// no recv
24 		ch <- 1 // L02           // <-------
25 	}()
26 	time.Sleep(10 * time.Millisecond)
27 }
28 
29 // TestUnbufferedRecvNoSend: recv blocks forever
30 func TestUnbufferedRecvNoSend(t *testing.T) {
31 	ch := make(chan int)
32 	// no send
33 	_ = <-ch // L01 or deadlock
34 }
35 
36 ...
```


## Replay
**Replaying was not run**.

