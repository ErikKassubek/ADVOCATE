# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestH2c
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/server.go:852
```go
841 ...
842 
843 	readMore func()
844 }
845 
846 // readFrames is the loop that reads incoming frames.
847 // It takes care to only read one frame at a time, blocking until the
848 // consumer is done with the frame.
849 // It's run on its own goroutine.
850 func (sc *serverConn) readFrames() {
851 	sc.srv.markNewGoroutine()
852 	gate := make(chan struct{})           // <-------
853 	gateDone := func() { gate <- struct{}{} }
854 	for {
855 		f, err := sc.framer.ReadFrame()
856 		select {
857 		case sc.readFrameCh <- readFrameResult{f, err, gateDone}:
858 		case <-sc.doneServing:
859 			return
860 		}
861 		select {
862 		case <-gate:
863 
864 ...
```


## Replay
**Replaying was not run**.

