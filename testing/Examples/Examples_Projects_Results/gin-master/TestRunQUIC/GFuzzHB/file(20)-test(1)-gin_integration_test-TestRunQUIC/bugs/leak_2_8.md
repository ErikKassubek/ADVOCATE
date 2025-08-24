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

