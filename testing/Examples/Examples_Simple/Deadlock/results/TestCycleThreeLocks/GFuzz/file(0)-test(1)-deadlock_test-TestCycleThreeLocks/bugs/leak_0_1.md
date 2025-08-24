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
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:42
```go
31 ...
32 
33 		x.Lock()
34 		time.Sleep(5 * time.Millisecond)
35 		y.Lock()
36 		y.Unlock()
37 		x.Unlock()
38 	}()
39 	go func() {
40 		y.Lock()
41 		time.Sleep(5 * time.Millisecond)
42 		z.Lock()           // <-------
43 		z.Unlock()
44 		y.Unlock()
45 	}()
46 	// main goroutine forms the third link:
47 	z.Lock()
48 	time.Sleep(5 * time.Millisecond)
49 	x.Lock() // deadlock on x
50 	x.Unlock()
51 	z.Unlock()
52 }
53 
54 ...
```


## Replay
**Replaying was not run**.

