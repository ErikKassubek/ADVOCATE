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
-> /home/ilian/Projects/go/pkg/mod/github.com/quic-go/quic-go@v0.52.0/server.go:334
```go
323 ...
324 
325 }
326 
327 // Accept returns connections that already completed the handshake.
328 // It is only valid if acceptEarlyConns is false.
329 func (s *baseServer) Accept(ctx context.Context) (Connection, error) {
330 	return s.accept(ctx)
331 }
332 
333 func (s *baseServer) accept(ctx context.Context) (quicConn, error) {
334 	select {           // <-------
335 	case <-ctx.Done():
336 		return nil, ctx.Err()
337 	case conn := <-s.connQueue:
338 		return conn, nil
339 	case <-s.stopAccepting:
340 		// first drain the queue
341 		select {
342 		case conn := <-s.connQueue:
343 			return conn, nil
344 		default:
345 
346 ...
```


## Replay
**Replaying was not run**.

