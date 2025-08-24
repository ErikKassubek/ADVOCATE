# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCycleThreeLocks
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:49
```go
38 ...
39 
40 		y.Lock()
41 		time.Sleep(5 * time.Millisecond)
42 		z.Lock()
43 		z.Unlock()
44 		y.Unlock()
45 	}()
46 	// main goroutine forms the third link:
47 	z.Lock()
48 	time.Sleep(5 * time.Millisecond)
49 	x.Lock() // deadlock on x           // <-------
50 	x.Unlock()
51 	z.Unlock()
52 }
53 
54 // 3) Channel‐based deadlock: send+recv missing
55 func TestChannelDeadlock(t *testing.T) {
56 	ch := make(chan int)
57 	// no corresponding receive
58 	ch <- 1
59 }
60 
61 ...
```


## Replay
**Replaying was not run**.

