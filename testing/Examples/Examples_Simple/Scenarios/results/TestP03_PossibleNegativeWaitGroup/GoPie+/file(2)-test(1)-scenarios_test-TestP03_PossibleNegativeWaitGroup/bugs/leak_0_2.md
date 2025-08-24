# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestP03_PossibleNegativeWaitGroup
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:86
```go
75 ...
76 
77 }
78 
79 func TestP03_PossibleNegativeWaitGroup(t *testing.T) {
80 	var wg sync.WaitGroup
81 	wg.Add(1)
82 	go func() {
83 		wg.Done()
84 		wg.Done() // P03
85 	}()
86 	wg.Wait()           // <-------
87 }
88 
89 // ---------------- L00–L10: Leaks ----------------
90 
91 func TestL00_UnknownLeak(t *testing.T) {
92 	done := make(chan struct{})
93 	go func() { <-done }() // never signaled → L00
94 	time.Sleep(20 * time.Millisecond)
95 }
96 
97 
98 ...
```


## Replay
**Replaying was not run**.

