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
-> /home/ilian/Projects/go/pkg/mod/github.com/quic-go/quic-go@v0.52.0/server.go:312
```go
301 ...
302 
303 			if bufferStillInUse := s.handlePacketImpl(p); !bufferStillInUse {
304 				p.buffer.Release()
305 			}
306 		}
307 	}
308 }
309 
310 func (s *baseServer) runSendQueue() {
311 	for {
312 		select {           // <-------
313 		case <-s.running:
314 			return
315 		case p := <-s.versionNegotiationQueue:
316 			s.maybeSendVersionNegotiationPacket(p)
317 		case p := <-s.invalidTokenQueue:
318 			s.maybeSendInvalidToken(p)
319 		case p := <-s.connectionRefusedQueue:
320 			s.sendConnectionRefused(p)
321 		case p := <-s.retryQueue:
322 			s.sendRetry(p)
323 
324 ...
```


## Replay
**Replaying was not run**.

