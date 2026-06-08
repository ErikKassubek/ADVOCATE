# Bug: A08 - Actual Cyclic Deadlock

During the execution, a deadlock was detected.
This means, there is are routine that are cyclicaly blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestGrpc1353
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/grpc/1353/grpc1353_test.go
- Trace: advocateTrace_29

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Select:
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/2448/cockroach2448_test.go#30
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/2448/cockroach2448_test.go#59
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/35073/cockroach35073_test.go#49
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/6181/cockroach6181_test.go#38
```go
27 ...
28 
29 	rdc.rangeCacheMu.RLock()
30 	fmt.Printf("lookup range descriptor: %s", rdc)
31 	rdc.rangeCacheMu.RUnlock()
32 	rdc.rangeCacheMu.Lock()
33 	rdc.rangeCacheMu.Unlock()
34 }
35 
36 func (rdc *rangeDescriptorCache) String() string {
37 	rdc.rangeCacheMu.RLock()
38 	defer rdc.rangeCacheMu.RUnlock()           // <-------
39 	return rdc.stringLocked()
40 }
41 
42 func (rdc *rangeDescriptorCache) stringLocked() string {
43 	return "something here"
44 }
45 
46 func doLookupWithToken(rc *rangeDescriptorCache) {
47 	rc.LookupRangeDescriptor()
48 }
49 
50 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/6181/cockroach6181_test.go#30
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/etcd/6873/etcd6873_test.go#39
```go
28 ...
29 
30 		for wb := range wbs.updatec {
31 			wbs.coalesce(wb)
32 		}
33 	}()
34 	return wbs
35 }
36 
37 func (wbs *watchBroadcasts) coalesce(wb *watchBroadcast) {
38 	wbs.mu.Lock()
39 	wbs.mu.Unlock()           // <-------
40 }
41 
42 func (wbs *watchBroadcasts) stop() {
43 	wbs.mu.Lock()
44 	defer wbs.mu.Unlock()
45 	close(wbs.updatec)
46 	<-wbs.donec
47 }
48 
49 func (wbs *watchBroadcasts) update(wb *watchBroadcast) {
50 
51 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/etcd/6873/etcd6873_test.go#47
```go
36 ...
37 
38 	wbs.mu.Lock()
39 	wbs.mu.Unlock()
40 }
41 
42 func (wbs *watchBroadcasts) stop() {
43 	wbs.mu.Lock()
44 	defer wbs.mu.Unlock()
45 	close(wbs.updatec)
46 	<-wbs.donec
47 }           // <-------
48 
49 func (wbs *watchBroadcasts) update(wb *watchBroadcast) {
50 	select {
51 	case wbs.updatec <- wb:
52 	default:
53 	}
54 }
55 
56 ///
57 /// G1						G2					G3
58 
59 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/grpc/1353/grpc1353_test.go#82
```go
71 ...
72 
73 }
74 
75 func (rr *roundRobin) watchAddrUpdates() {
76 	rr.mu.Lock()
77 	defer rr.mu.Unlock()
78 	rr.addrCh <- true // Block here
79 }
80 
81 func (rr *roundRobin) down() {
82 	rr.mu.Lock() // Block here           // <-------
83 	defer rr.mu.Unlock()
84 }
85 
86 type addrConn struct {
87 	mu   sync.Mutex
88 	down func()
89 }
90 
91 func (ac *addrConn) tearDown() {
92 	ac.mu.Lock()
93 
94 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/grpc/1353/grpc1353_test.go#78
```go
67 ...
68 
69 	defer rr.mu.Unlock()
70 	if rr.addrCh != nil {
71 		close(rr.addrCh)
72 	}
73 }
74 
75 func (rr *roundRobin) watchAddrUpdates() {
76 	rr.mu.Lock()
77 	defer rr.mu.Unlock()
78 	rr.addrCh <- true // Block here           // <-------
79 }
80 
81 func (rr *roundRobin) down() {
82 	rr.mu.Lock() // Block here
83 	defer rr.mu.Unlock()
84 }
85 
86 type addrConn struct {
87 	mu   sync.Mutex
88 	down func()
89 
90 ...
```


## Replay
**Replaying was not run**.

