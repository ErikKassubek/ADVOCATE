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
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/transport.go:1473
```go
1462 ...
1463 
1464 			return errRequestCanceled
1465 		case <-ctx.Done():
1466 			return ctx.Err()
1467 		case <-cc.seenSettingsChan:
1468 			if !cc.extendedConnectAllowed {
1469 				return errExtendedConnectNotSupported
1470 			}
1471 		}
1472 	}
1473 	select {           // <-------
1474 	case cc.reqHeaderMu <- struct{}{}:
1475 	case <-cs.reqCancel:
1476 		return errRequestCanceled
1477 	case <-ctx.Done():
1478 		return ctx.Err()
1479 	}
1480 
1481 	cc.mu.Lock()
1482 	if cc.idleTimer != nil {
1483 		cc.idleTimer.Stop()
1484 
1485 ...
```


## Replay
**Replaying was not run**.

