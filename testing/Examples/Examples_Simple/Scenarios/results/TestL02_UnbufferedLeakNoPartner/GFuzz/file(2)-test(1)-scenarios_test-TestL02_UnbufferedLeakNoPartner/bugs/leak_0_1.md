# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestL02_UnbufferedLeakNoPartner
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:108
```go
97 ...
98 
99 	go func() {
100 		time.Sleep(10 * time.Millisecond)
101 		_ = <-ch // partner exists
102 	}()
103 	ch <- 1 // L01 (racing)
104 }
105 
106 func TestL02_UnbufferedLeakNoPartner(t *testing.T) {
107 	ch := make(chan int)
108 	ch <- 1 // L02 → no receiver           // <-------
109 }
110 
111 func TestL03_BufferedLeakWithPartner(t *testing.T) {
112 	ch := make(chan int, 1)
113 	ch <- 1
114 	go func() {
115 		time.Sleep(10 * time.Millisecond)
116 		_ = <-ch // L03
117 	}()
118 }
119 
120 ...
```


## Replay
**Replaying was not run**.

