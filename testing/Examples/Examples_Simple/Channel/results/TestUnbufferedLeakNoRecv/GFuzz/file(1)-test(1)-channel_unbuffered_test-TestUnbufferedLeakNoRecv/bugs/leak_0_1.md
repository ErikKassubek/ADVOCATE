# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestUnbufferedLeakNoRecv
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Channel/channel_unbuffered_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
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

