# Bug: A08 - Actual Cyclic Deadlock

During the execution, a deadlock was detected.
This means, there is are routine that are cyclicaly blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMoby4951
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/moby/4951/moby4951_test.go
- Trace: advocateTrace_65

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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/grpc/1353/grpc1353_test.go#83
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/grpc/1353/grpc1353_test.go#79
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/grpc/795/grpc795_test.go#15
```go
4 ...
5 
6 )
7 
8 type Server struct {
9 	mu    sync.Mutex
10 	drain bool
11 }
12 
13 func (s *Server) GracefulStop() {
14 	s.mu.Lock()
15 	if s.drain == true {           // <-------
16 		s.mu.Lock()
17 		return
18 	}
19 	s.drain = true
20 } // Missing Unlock
21 
22 func (s *Server) Serve() {
23 	s.mu.Lock()
24 	s.mu.Unlock()
25 }
26 
27 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/grpc/795/grpc795_test.go#24
```go
13 ...
14 
15 	if s.drain == true {
16 		s.mu.Lock()
17 		return
18 	}
19 	s.drain = true
20 } // Missing Unlock
21 
22 func (s *Server) Serve() {
23 	s.mu.Lock()
24 	s.mu.Unlock()           // <-------
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
35 
36 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/istio/16224/istio16224_test.go#106
```go
95 ...
96 
97 	go controller.Run(stop)
98 
99 	controller.Create()
100 
101 	lock.Lock()
102 	lock.Unlock()
103 	<-done
104 
105 	close(stop)
106 }           // <-------
107 
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/istio/16224/istio16224_test.go#98
```go
87 ...
88 
89 	lock := sync.Mutex{}
90 	controller.RegisterEventHandler(func(event Event) {
91 		lock.Lock()
92 		defer lock.Unlock()
93 		done <- true
94 	})
95 
96 	stop := make(chan struct{})
97 	go controller.Run(stop)
98            // <-------
99 	controller.Create()
100 
101 	lock.Lock()
102 	lock.Unlock()
103 	<-done
104 
105 	close(stop)
106 }
107 
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/istio/17860/istio17860_test.go#71
```go
60 ...
61 
62 	a.waitUntilLive()
63 	a.currentEpoch++
64 	a.activeEpochs[a.currentEpoch] = struct{}{}
65 
66 	go a.runWait(a.currentEpoch)
67 }
68 
69 func (a *agent) runWait(epoch int) {
70 	a.statusCh <- exitStatus(epoch)
71 }           // <-------
72 
73 func (a *agent) waitUntilLive() {
74 	if len(a.activeEpochs) == 0 {
75 		return
76 	}
77 
78 	interval := time.NewTicker(30 * time.Nanosecond)
79 	timer := time.NewTimer(100 * time.Nanosecond)
80 	defer func() {
81 		interval.Stop()
82 
83 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/hugo/3251/hugo3251_test.go#21
```go
10 ...
11 
12 )
13 
14 type remoteLock struct {
15 	sync.RWMutex
16 	m map[string]*sync.Mutex
17 }
18 
19 func (l *remoteLock) URLLock(url string) {
20 	l.Lock()
21 	if _, ok := l.m[url]; !ok {           // <-------
22 		l.m[url] = &sync.Mutex{}
23 	}
24 	l.m[url].Lock()
25 	l.Unlock()
26 }
27 
28 func (l *remoteLock) URLUnlock(url string) {
29 	l.RLock()
30 	defer l.RUnlock()
31 	if um, ok := l.m[url]; ok {
32 
33 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/moby/4951/moby4951_test.go#33
```go
22 ...
23 
24 	nrDeletedDevices int
25 }
26 
27 func (devices *DeviceSet) DeleteDevice(hash string) {
28 	devices.Lock()
29 	defer devices.Unlock()
30 
31 	info := devices.lookupDevice(hash)
32 
33 	info.lock.Lock()           // <-------
34 	defer info.lock.Unlock()
35 
36 	devices.deleteDevice(info)
37 }
38 
39 func (devices *DeviceSet) lookupDevice(hash string) *DevInfo {
40 	existing, ok := devices.infos[hash]
41 	if !ok {
42 		return nil
43 	}
44 
45 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/moby/4951/moby4951_test.go#55
```go
44 ...
45 
46 
47 func (devices *DeviceSet) deleteDevice(info *DevInfo) {
48 	devices.removeDeviceAndWait(info.Name())
49 }
50 
51 func (devices *DeviceSet) removeDeviceAndWait(devname string) {
52 	/// remove devices by devname
53 	devices.Unlock()
54 	time.Sleep(300 * time.Nanosecond)
55 	devices.Lock()           // <-------
56 }
57 
58 type DevInfo struct {
59 	lock sync.Mutex
60 	name string
61 }
62 
63 func (info *DevInfo) Name() string {
64 	return info.name
65 }
66 
67 ...
```


## Replay
**Replaying was not run**.

