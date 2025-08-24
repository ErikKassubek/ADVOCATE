# Leak: Leak on select with possible partner

The analyzer detected a Leak on a select with a possible partner.
A Leak on a select is a situation, where a select is still blocking at the end of the program.
The partner is a corresponding send or receive operation, which communicated with another operation, but could communicated with the stuck operation instead, resolving the leak.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestH2c
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Select:
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/server.go:987
```go
976 ...
977 
978 	go sc.readFrames() // closed by defer sc.conn.Close above
979 
980 	settingsTimer := sc.srv.afterFunc(firstSettingsTimeout, sc.onSettingsTimer)
981 	defer settingsTimer.Stop()
982 
983 	lastFrameTime := sc.srv.now()
984 	loopNum := 0
985 	for {
986 		loopNum++
987 		select {           // <-------
988 		case wr := <-sc.wantWriteFrameCh:
989 			if se, ok := wr.write.(StreamError); ok {
990 				sc.resetStream(se)
991 				break
992 			}
993 			sc.writeFrame(wr)
994 		case res := <-sc.wroteFrameCh:
995 			sc.wroteFrame(res)
996 		case res := <-sc.readFrameCh:
997 			lastFrameTime = sc.srv.now()
998 
999 ...
```


###  Channel: Send
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/server.go:856
```go
845 ...
846 
847 // It takes care to only read one frame at a time, blocking until the
848 // consumer is done with the frame.
849 // It's run on its own goroutine.
850 func (sc *serverConn) readFrames() {
851 	sc.srv.markNewGoroutine()
852 	gate := make(chan struct{})
853 	gateDone := func() { gate <- struct{}{} }
854 	for {
855 		f, err := sc.framer.ReadFrame()
856 		select {           // <-------
857 		case sc.readFrameCh <- readFrameResult{f, err, gateDone}:
858 		case <-sc.doneServing:
859 			return
860 		}
861 		select {
862 		case <-gate:
863 		case <-sc.doneServing:
864 			return
865 		}
866 		if terminalReadFrameError(err) {
867 
868 ...
```


## Replay
The analyzer found a leak in the recorded trace.
The analyzer found a way to resolve the leak, meaning the leak should not reappear in the rewritten trace.

**Replaying failed**.

It exited with the following code: 10

The replay got stuck during the execution.
The main routine has already finished, but the trace still contains not executed operations.
This can be caused by a stuck replay.
Possible causes are:
    - The program was altered between recording and replay
    - The program execution path is not deterministic, e.g. its execution path is determined by a random number
    - The program execution path depends on the order of not tracked operations
    - The program execution depends on outside input, that was not exactly reproduced
	 - The program encountered a deadlock earlier in the trace than expected

