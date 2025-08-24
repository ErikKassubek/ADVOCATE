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
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/server.go:891
```go
880 ...
881 
882 // serverConn.
883 func (sc *serverConn) writeFrameAsync(wr FrameWriteRequest, wd *writeData) {
884 	sc.srv.markNewGoroutine()
885 	var err error
886 	if wd == nil {
887 		err = wr.write.writeFrame(sc)
888 	} else {
889 		err = sc.framer.endWrite()
890 	}
891 	sc.wroteFrameCh <- frameWriteResult{wr: wr, err: err}           // <-------
892 }
893 
894 func (sc *serverConn) closeAllStreamsOnConnClose() {
895 	sc.serveG.check()
896 	for _, st := range sc.streams {
897 		sc.closeStream(st, errClientDisconnected)
898 	}
899 }
900 
901 func (sc *serverConn) stopShutdownTimer() {
902 
903 ...
```


## Replay
**Replaying was not run**.

