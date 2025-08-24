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


## Replay
**Replaying was not run**.

