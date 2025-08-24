# Leak: Leak on unbuffered channel with possible partner

The analyzer detected a Leak on an unbuffered channel with a possible partner.
A Leak on an unbuffered channel is a situation, where a unbuffered channel is still blocking at the end of the program.
The partner is a corresponding send or receive operation, which communicated with another operation, but could communicated with the stuck operation instead, resolving the deadlock.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestL06_LeakOnSelectWithPartner
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Channel: Send
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


###  Channel: Receive
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:136
```go
125 ...
126 
127 	ch <- 1 // L05 → nil channel send (blocks forever)
128 }
129 
130 func TestL06_LeakOnSelectWithPartner(t *testing.T) {
131 	ch1 := make(chan int)
132 	go func() {
133 		time.Sleep(10 * time.Millisecond)
134 		ch1 <- 42
135 	}()
136 	select {           // <-------
137 	case <-ch1:
138 	case <-time.After(50 * time.Millisecond):
139 	}
140 }
141 
142 func TestL07_LeakOnSelectWithoutPartner(t *testing.T) {
143 	var ch chan int // nil channel
144 	select {
145 	case <-ch:
146 	case <-time.After(50 * time.Millisecond):
147 
148 ...
```


## Replay
The analyzer found a leak in the recorded trace.
The analyzer found a way to resolve the leak, meaning the leak should not reappear in the rewritten trace.

**Replaying was successful**.

It exited with the following code: 20

The replay was able to get the leaking unbuffered channel or select unstuck.

