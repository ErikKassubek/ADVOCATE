# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestL00_UnknownLeak
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:93
```go
82 ...
83 
84 		wg.Done() // P03
85 	}()
86 	wg.Wait()
87 }
88 
89 // ---------------- L00–L10: Leaks ----------------
90 
91 func TestL00_UnknownLeak(t *testing.T) {
92 	done := make(chan struct{})
93 	go func() { <-done }() // never signaled → L00           // <-------
94 	time.Sleep(20 * time.Millisecond)
95 }
96 
97 func TestL01_UnbufferedLeakWithPartner(t *testing.T) {
98 	ch := make(chan int)
99 	go func() {
100 		time.Sleep(10 * time.Millisecond)
101 		_ = <-ch // partner exists
102 	}()
103 	ch <- 1 // L01 (racing)
104 
105 ...
```


## Replay
**Replaying was not run**.

