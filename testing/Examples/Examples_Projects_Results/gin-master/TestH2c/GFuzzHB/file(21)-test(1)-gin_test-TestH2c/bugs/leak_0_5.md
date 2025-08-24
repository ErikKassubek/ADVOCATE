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
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/transport.go:2775
```go
2764 ...
2765 
2766 		// Close cs.bufPipe and cs.peerClosed with cc.mu held to avoid a
2767 		// race condition: The caller can read io.EOF from Response.Body
2768 		// and close the body before we close cs.peerClosed, causing
2769 		// cleanupWriteRequest to send a RST_STREAM.
2770 		rl.cc.mu.Lock()
2771 		defer rl.cc.mu.Unlock()
2772 		cs.bufPipe.closeWithErrorAndCode(io.EOF, cs.copyTrailers)
2773 		close(cs.peerClosed)
2774 	}
2775 }           // <-------
2776 
2777 func (rl *clientConnReadLoop) endStreamError(cs *clientStream, err error) {
2778 	cs.readAborted = true
2779 	cs.abortStream(err)
2780 }
2781 
2782 // Constants passed to streamByID for documentation purposes.
2783 const (
2784 	headerOrDataFrame    = true
2785 	notHeaderOrDataFrame = false
2786 
2787 ...
```


## Replay
**Replaying was not run**.

