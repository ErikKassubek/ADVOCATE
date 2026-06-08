# Bug: A08 - Actual Cyclic Deadlock

During the execution, a deadlock was detected.
This means, there is are routine that are cyclicaly blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach7504
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/cockroach/7504/cockroach7504_test.go
- Trace: advocateTrace_18

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/cockroach/10790/cockroach10790_test.go#63
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/cockroach/2448/cockroach2448_test.go#59
```go
48 ...
49 
50 }
51 func (s *state) handleWriteResponse() {
52 	s.processCommittedEntry()
53 }
54 
55 func (s *state) processCommittedEntry() {
56 	s.sendEvent(&EventMembershipChangeCommitted{
57 		Callback: func() {
58 			select {
59 			case s.callbackChan <- func() { // Waiting for callbackChan consumption           // <-------
60 				time.Sleep(time.Nanosecond)
61 			}:
62 			case <-s.stopper.ShouldStop():
63 			}
64 		},
65 	})
66 }
67 
68 type Store struct {
69 	multiraft *MultiRaft
70 
71 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/cockroach/2448/cockroach2448_test.go#30
```go
19 ...
20 
21 	Events       chan interface{}
22 	callbackChan chan func()
23 }
24 
25 // sendEvent can be invoked many times
26 func (m *MultiRaft) sendEvent(event interface{}) {
27 	/// FIX:
28 	/// Let event append a event queue instead of pending here
29 	select {
30 	case m.Events <- event: // Waiting for events consumption           // <-------
31 	case <-m.stopper.ShouldStop():
32 	}
33 }
34 
35 type state struct {
36 	*MultiRaft
37 }
38 
39 func (s *state) start() {
40 	for {
41 
42 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/cockroach/35073/cockroach35073_test.go#49
```go
38 ...
39 
40 	}
41 }
42 
43 func (rc *RowChannel) Push() ConsumerStatus {
44 	consumerStatus := ConsumerStatus(
45 		atomic.LoadUint32((*uint32)(&rc.consumerStatus)))
46 	switch consumerStatus {
47 	case NeedMoreRows:
48 		rc.dataChan <- RowChannelMsg(0)
49 	case DrainRequested:           // <-------
50 	case ConsumerClosed:
51 	}
52 	return consumerStatus
53 }
54 
55 func (rc *RowChannel) InitWithNumSenders() {
56 	rc.initWithBufSizeAndNumSenders(rowChannelBufSize)
57 }
58 
59 func (rc *RowChannel) initWithBufSizeAndNumSenders(chanBufSize int) {
60 
61 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/cockroach/6181/cockroach6181_test.go#30
```go
19 ...
20 
21 	return &testDescriptorDB{&rangeDescriptorCache{}}
22 }
23 
24 type rangeDescriptorCache struct {
25 	rangeCacheMu sync.RWMutex
26 }
27 
28 func (rdc *rangeDescriptorCache) LookupRangeDescriptor() {
29 	rdc.rangeCacheMu.RLock()
30 	fmt.Printf("lookup range descriptor: %s", rdc)           // <-------
31 	rdc.rangeCacheMu.RUnlock()
32 	rdc.rangeCacheMu.Lock()
33 	rdc.rangeCacheMu.Unlock()
34 }
35 
36 func (rdc *rangeDescriptorCache) String() string {
37 	rdc.rangeCacheMu.RLock()
38 	defer rdc.rangeCacheMu.RUnlock()
39 	return rdc.stringLocked()
40 }
41 
42 ...
```


## Replay
**Replaying was not run**.

