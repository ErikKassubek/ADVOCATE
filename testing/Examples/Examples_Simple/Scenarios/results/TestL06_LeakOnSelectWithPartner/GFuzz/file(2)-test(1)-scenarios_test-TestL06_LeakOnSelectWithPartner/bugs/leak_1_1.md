# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestL06_LeakOnSelectWithPartner
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:134
```go
123 ...
124 
125 func TestL05_LeakOnNilChan(t *testing.T) {
126 	var ch chan int
127 	ch <- 1 // L05 → nil channel send (blocks forever)
128 }
129 
130 func TestL06_LeakOnSelectWithPartner(t *testing.T) {
131 	ch1 := make(chan int)
132 	go func() {
133 		time.Sleep(10 * time.Millisecond)
134 		ch1 <- 42           // <-------
135 	}()
136 	select {
137 	case <-ch1:
138 	case <-time.After(50 * time.Millisecond):
139 	}
140 }
141 
142 func TestL07_LeakOnSelectWithoutPartner(t *testing.T) {
143 	var ch chan int // nil channel
144 	select {
145 
146 ...
```


## Replay
**Replaying was not run**.

