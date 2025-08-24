# Leak: Leak on unbuffered channel without possible partner

The analyzer detected a Leak on an unbuffered channel without a possible partner.
A Leak on an unbuffered channel is a situation, where a unbuffered channel is still blocking at the end of the program.
The analyzer could not find a partner for the stuck operation, which would resolve the leak.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestChannelDeadlock
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Deadlock/deadlock_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Channel: Send
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

