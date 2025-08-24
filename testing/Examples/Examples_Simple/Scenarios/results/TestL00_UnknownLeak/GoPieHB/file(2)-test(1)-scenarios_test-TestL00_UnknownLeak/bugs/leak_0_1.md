# Leak: Leak on unbuffered channel without possible partner

The analyzer detected a Leak on an unbuffered channel without a possible partner.
A Leak on an unbuffered channel is a situation, where a unbuffered channel is still blocking at the end of the program.
The analyzer could not find a partner for the stuck operation, which would resolve the leak.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestL00_UnknownLeak
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Channel: Receive
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

