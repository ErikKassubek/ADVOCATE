# Leak: Leak on select without possible partner

The analyzer detected a Leak on a select without a possible partner.
A Leak on a select is a situation, where a select is still blocking at the end of the program.
The analyzer could not find a partner for the stuck operation, which would resolve the leak.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRunQUIC
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_integration_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Select:
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

