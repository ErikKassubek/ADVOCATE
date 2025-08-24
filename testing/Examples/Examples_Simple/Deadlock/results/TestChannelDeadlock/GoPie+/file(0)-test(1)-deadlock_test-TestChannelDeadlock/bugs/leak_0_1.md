# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestChannelDeadlock
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go:58
```go
47 ...
48 
49 	x.Lock() // deadlock on x
50 	x.Unlock()
51 	z.Unlock()
52 }
53 
54 // 3) Channel‐based deadlock: send+recv missing
55 func TestChannelDeadlock(t *testing.T) {
56 	ch := make(chan int)
57 	// no corresponding receive
58 	ch <- 1           // <-------
59 }
60 
61 // 4) WaitGroup deadlock: missing Done
62 func TestWaitGroupDeadlock(t *testing.T) {
63 	var wg sync.WaitGroup
64 	wg.Add(1)
65 	go func() {
66 		// forgot wg.Done()
67 		time.Sleep(20 * time.Millisecond)
68 	}()
69 
70 ...
```


## Replay
**Replaying was not run**.

