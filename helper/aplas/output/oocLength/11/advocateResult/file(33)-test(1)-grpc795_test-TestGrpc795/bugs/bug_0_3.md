# Bug: A08 - Actual Cyclic Deadlock

During the execution, a deadlock was detected.
This means, there is are routine that are cyclicaly blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestGrpc795
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/grpc/795/grpc795_test.go
- Trace: advocateTrace_34

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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/etcd/6873/etcd6873_test.go#47
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/etcd/6873/etcd6873_test.go#39
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/grpc/1353/grpc1353_test.go#83
```go
72 ...
73 
74 
75 func (rr *roundRobin) watchAddrUpdates() {
76 	rr.mu.Lock()
77 	defer rr.mu.Unlock()
78 	rr.addrCh <- true // Block here
79 }
80 
81 func (rr *roundRobin) down() {
82 	rr.mu.Lock() // Block here
83 	defer rr.mu.Unlock()           // <-------
84 }
85 
86 type addrConn struct {
87 	mu   sync.Mutex
88 	down func()
89 }
90 
91 func (ac *addrConn) tearDown() {
92 	ac.mu.Lock()
93 	defer ac.mu.Unlock()
94 
95 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/grpc/1353/grpc1353_test.go#79
```go
68 ...
69 
70 	if rr.addrCh != nil {
71 		close(rr.addrCh)
72 	}
73 }
74 
75 func (rr *roundRobin) watchAddrUpdates() {
76 	rr.mu.Lock()
77 	defer rr.mu.Unlock()
78 	rr.addrCh <- true // Block here
79 }           // <-------
80 
81 func (rr *roundRobin) down() {
82 	rr.mu.Lock() // Block here
83 	defer rr.mu.Unlock()
84 }
85 
86 type addrConn struct {
87 	mu   sync.Mutex
88 	down func()
89 }
90 
91 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/grpc/3017/grpc3017_test.go#31
```go
20 ...
21 
22 	subConnToAddr map[SubConn]Address
23 }
24 
25 func (ccc *lbCacheClientConn) NewSubConn(addrs []Address) SubConn {
26 	if len(addrs) != 1 {
27 		return SubConn(1)
28 	}
29 	addrWithoutMD := addrs[0]
30 	ccc.mu.Lock()
31 	defer ccc.mu.Unlock()           // <-------
32 	if entry, ok := ccc.subConnCache[addrWithoutMD]; ok {
33 		entry.cancel()
34 		delete(ccc.subConnCache, addrWithoutMD)
35 		return entry.sc
36 	}
37 	scNew := SubConn(1)
38 	ccc.subConnToAddr[scNew] = addrWithoutMD
39 	return scNew
40 }
41 
42 
43 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/grpc/3017/grpc3017_test.go#64
```go
53 ...
54 
55 	}
56 
57 	entry := &subConnCacheEntry{
58 		sc: sc,
59 	}
60 	ccc.subConnCache[addr] = entry
61 
62 	timer := time.AfterFunc(ccc.timeout, func() {
63 		ccc.mu.Lock()
64 		if entry.abortDeleting {           // <-------
65 			return // Missing unlock
66 		}
67 		delete(ccc.subConnToAddr, sc)
68 		delete(ccc.subConnCache, addr)
69 		ccc.mu.Unlock()
70 	})
71 
72 	entry.cancel = func() {
73 		if !timer.Stop() {
74 			entry.abortDeleting = true
75 
76 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/grpc/795/grpc795_test.go#14
```go
3 ...
4 
5 	"testing"
6 )
7 
8 type Server struct {
9 	mu    sync.Mutex
10 	drain bool
11 }
12 
13 func (s *Server) GracefulStop() {
14 	s.mu.Lock()           // <-------
15 	if s.drain == true {
16 		s.mu.Lock()
17 		return
18 	}
19 	s.drain = true
20 } // Missing Unlock
21 
22 func (s *Server) Serve() {
23 	s.mu.Lock()
24 	s.mu.Unlock()
25 
26 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/11/blocking/grpc/795/grpc795_test.go#23
```go
12 ...
13 
14 	s.mu.Lock()
15 	if s.drain == true {
16 		s.mu.Lock()
17 		return
18 	}
19 	s.drain = true
20 } // Missing Unlock
21 
22 func (s *Server) Serve() {
23 	s.mu.Lock()           // <-------
24 	s.mu.Unlock()
25 }
26 
27 func NewServer() *Server {
28 	return &Server{}
29 }
30 
31 type test struct {
32 	srv *Server
33 }
34 
35 ...
```


## Replay
**Replaying was not run**.

