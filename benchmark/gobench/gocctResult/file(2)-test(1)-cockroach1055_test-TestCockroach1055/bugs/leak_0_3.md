# Leak: L00 - Leak

The analyzer detected a leak.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
A Leak could potentially resolve itself, if the program would run longer.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach1055
- File: /home/erik/Arbeit/Advocate/benchmark/gobench/cockroach/1055/cockroach1055_test.go
- Trace: gocctTrace_2

## Bug Elements

The elements involved in the found leak are located at the following positions:

###  Unknown element type

-> /home/erik/Arbeit/Advocate/goPatch/src/context/context.go#477

```go...
467 	}
468 	return nil
469 }
470 
471 // propagateCancel arranges for child to be canceled when parent is.
472 // It sets the parent context of cancelCtx.
473 func (c *cancelCtx) propagateCancel(parent Context, child canceler) {
474 	c.Context = parent
475 
476 	done := parent.Done()
477 	if done == nil {                    // <================= 
478 		return // parent is never canceled
479 	}
480 
481 	select {
482 	case <-done:
483 		// parent is already canceled
484 		child.cancel(false, parent.Err(), Cause(parent))
485 		return
486 	default:
487 	}
...
```

