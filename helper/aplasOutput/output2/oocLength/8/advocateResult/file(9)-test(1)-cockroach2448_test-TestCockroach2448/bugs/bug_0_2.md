# Bug: A08 - Actual Cyclic Deadlock

During the execution, a deadlock was detected.
This means, there is are routine that are cyclicaly blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach2448
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/2448/cockroach2448_test.go
- Trace: advocateTrace_10

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/10790/cockroach10790_test.go#63
```go
52 ...
53 
54 func (r *Replica) beginCmds(ctx context.Context) {
55 	ctxDone := ctx.Done()
56 	for _, ch := range r.chans {
57 		select {
58 		case <-ch:
59 		case <-ctxDone:
60 			go func() {
61 				for _, ch := range r.chans {
62 					<-ch
63 				}           // <-------
64 			}()
65 		}
66 	}
67 }
68 
69 /// helper goroutine, not present in the real bug.
70 func (r *Replica) sendChans(ctx context.Context) {
71 	for _, ch := range r.chans {
72 		select {
73 		case ch <- true:
74 
75 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/2448/cockroach2448_test.go#58
```go
47 ...
48 
49 	}
50 }
51 func (s *state) handleWriteResponse() {
52 	s.processCommittedEntry()
53 }
54 
55 func (s *state) processCommittedEntry() {
56 	s.sendEvent(&EventMembershipChangeCommitted{
57 		Callback: func() {
58 			select {           // <-------
59 			case s.callbackChan <- func() { // Waiting for callbackChan consumption
60 				time.Sleep(time.Nanosecond)
61 			}:
62 			case <-s.stopper.ShouldStop():
63 			}
64 		},
65 	})
66 }
67 
68 type Store struct {
69 
70 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/2448/cockroach2448_test.go#29
```go
18 ...
19 
20 	stopper      *Stopper
21 	Events       chan interface{}
22 	callbackChan chan func()
23 }
24 
25 // sendEvent can be invoked many times
26 func (m *MultiRaft) sendEvent(event interface{}) {
27 	/// FIX:
28 	/// Let event append a event queue instead of pending here
29 	select {           // <-------
30 	case m.Events <- event: // Waiting for events consumption
31 	case <-m.stopper.ShouldStop():
32 	}
33 }
34 
35 type state struct {
36 	*MultiRaft
37 }
38 
39 func (s *state) start() {
40 
41 ...
```


## Replay
**Replaying was not run**.

