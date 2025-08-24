# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestH2c
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/transport.go:1570
```go
1559 ...
1560 
1561 		timer := cc.t.newTimer(d)
1562 		defer timer.Stop()
1563 		respHeaderTimer = timer.C()
1564 		respHeaderRecv = cs.respHeaderRecv
1565 	}
1566 	// Wait until the peer half-closes its end of the stream,
1567 	// or until the request is aborted (via context, error, or otherwise),
1568 	// whichever comes first.
1569 	for {
1570 		select {           // <-------
1571 		case <-cs.peerClosed:
1572 			return nil
1573 		case <-respHeaderTimer:
1574 			return errTimeout
1575 		case <-respHeaderRecv:
1576 			respHeaderRecv = nil
1577 			respHeaderTimer = nil // keep waiting for END_STREAM
1578 		case <-cs.abort:
1579 			return cs.abortErr
1580 		case <-ctx.Done():
1581 
1582 ...
```


###  Channel: Close
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/transport.go:2773
```go
2762 ...
2763 
2764 	if !cs.readClosed {
2765 		cs.readClosed = true
2766 		// Close cs.bufPipe and cs.peerClosed with cc.mu held to avoid a
2767 		// race condition: The caller can read io.EOF from Response.Body
2768 		// and close the body before we close cs.peerClosed, causing
2769 		// cleanupWriteRequest to send a RST_STREAM.
2770 		rl.cc.mu.Lock()
2771 		defer rl.cc.mu.Unlock()
2772 		cs.bufPipe.closeWithErrorAndCode(io.EOF, cs.copyTrailers)
2773 		close(cs.peerClosed)           // <-------
2774 	}
2775 }
2776 
2777 func (rl *clientConnReadLoop) endStreamError(cs *clientStream, err error) {
2778 	cs.readAborted = true
2779 	cs.abortStream(err)
2780 }
2781 
2782 // Constants passed to streamByID for documentation purposes.
2783 const (
2784 
2785 ...
```


## Replay
**Replaying was not run**.

