# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRunQUIC
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_integration_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/go/pkg/mod/github.com/quic-go/quic-go@v0.52.0/server.go:299
```go
288 ...
289 
290 
291 func (s *baseServer) run() {
292 	defer close(s.running)
293 	for {
294 		select {
295 		case <-s.errorChan:
296 			return
297 		default:
298 		}
299 		select {           // <-------
300 		case <-s.errorChan:
301 			return
302 		case p := <-s.receivedPackets:
303 			if bufferStillInUse := s.handlePacketImpl(p); !bufferStillInUse {
304 				p.buffer.Release()
305 			}
306 		}
307 	}
308 }
309 
310 
311 ...
```


## Replay
**Replaying was not run**.

