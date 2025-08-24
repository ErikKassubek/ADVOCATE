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
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/transport.go:1596
```go
1585 ...
1586 
1587 
1588 func (cs *clientStream) encodeAndWriteHeaders(req *http.Request) error {
1589 	cc := cs.cc
1590 	ctx := cs.ctx
1591 
1592 	cc.wmu.Lock()
1593 	defer cc.wmu.Unlock()
1594 
1595 	// If the request was canceled while waiting for cc.mu, just quit.
1596 	select {           // <-------
1597 	case <-cs.abort:
1598 		return cs.abortErr
1599 	case <-ctx.Done():
1600 		return ctx.Err()
1601 	case <-cs.reqCancel:
1602 		return errRequestCanceled
1603 	default:
1604 	}
1605 
1606 	// Encode headers.
1607 
1608 ...
```


## Replay
**Replaying was not run**.

