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
-> /home/ilian/Projects/go/pkg/mod/github.com/quic-go/quic-go@v0.52.0/transport.go:449
```go
438 ...
439 
440 func (t *Transport) WriteTo(b []byte, addr net.Addr) (int, error) {
441 	if err := t.init(false); err != nil {
442 		return 0, err
443 	}
444 	return t.conn.WritePacket(b, addr, nil, 0, protocol.ECNUnsupported)
445 }
446 
447 func (t *Transport) runSendQueue() {
448 	for {
449 		select {           // <-------
450 		case <-t.listening:
451 			return
452 		case p := <-t.closeQueue:
453 			t.conn.WritePacket(p.payload, p.addr, p.info.OOB(), 0, protocol.ECNUnsupported)
454 		case p := <-t.statelessResetQueue:
455 			t.sendStatelessReset(p)
456 		}
457 	}
458 }
459 
460 
461 ...
```


## Replay
**Replaying was not run**.

