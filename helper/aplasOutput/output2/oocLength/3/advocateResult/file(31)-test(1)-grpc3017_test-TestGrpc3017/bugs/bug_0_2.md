# Bug: A07 - Actual Non-Cyclic Blocking Bug

During the execution, a blocking bug was detected.
This means, there is a routine that is blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestGrpc3017
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/grpc/3017/grpc3017_test.go
- Trace: advocateTrace_32

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/1055/cockroach1055_test.go#83
```go
72 ...
73 
74 		s := stoppers[i]
75 		s.AddWorker()
76 		go func() {
77 			s.StartTask()
78 			<-s.ShouldStop()
79 			s.SetStopped()
80 		}()
81 	}
82 
83 	done := make(chan struct{})           // <-------
84 	go func() {
85 		for _, s := range stoppers {
86 			s.Quiesce()
87 		}
88 		for _, s := range stoppers {
89 			s.Stop()
90 		}
91 		close(done)
92 	}()
93 
94 
95 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/1055/cockroach1055_test.go#99


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/1055/cockroach1055_test.go#47
```go
36 ...
37 
38 	s.drain.Wait()
39 	s.draining = 0
40 }
41 
42 func (s *Stopper) Stop() {
43 	s.mu.Lock()
44 	defer s.mu.Unlock()
45 	atomic.StoreInt32(&s.draining, 1)
46 	s.drain.Wait()
47 	close(s.stopper)           // <-------
48 	s.stop.Wait()
49 }
50 
51 func (s *Stopper) StartTask() bool {
52 	if atomic.LoadInt32(&s.draining) == 0 {
53 		s.mu.Lock()
54 		defer s.mu.Unlock()
55 		s.drain.Add(1)
56 		return true
57 	}
58 
59 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/13197/cockroach13197_test.go#36
```go
25 ...
26 
27 }
28 
29 type Tx struct {
30 	cancel context.CancelFunc
31 	ctx    context.Context
32 }
33 
34 func (tx *Tx) awaitDone() {
35 	<-tx.ctx.Done()
36 }           // <-------
37 
38 func (tx *Tx) Rollback() {
39 	tx.rollback()
40 }
41 
42 func (tx *Tx) rollback() {
43 	tx.close()
44 }
45 
46 func (tx *Tx) close() {
47 
48 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/13755/cockroach13755_test.go#30
```go
19 ...
20 
21 }
22 
23 func (rs *Rows) initContextClose(ctx context.Context) {
24 	ctx, rs.cancel = context.WithCancel(ctx)
25 	go rs.awaitDone(ctx)
26 }
27 
28 func (rs *Rows) awaitDone(ctx context.Context) {
29 	<-ctx.Done()
30 	rs.close(ctx.Err())           // <-------
31 }
32 
33 func (rs *Rows) close(err error) {
34 	// rs.cancel()
35 }
36 
37 /// G1 						G2
38 /// initContextClose()
39 /// 						awaitDone()
40 /// 						<-tx.ctx.Done()
41 
42 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/24808/cockroach24808_test.go#50
```go
39 ...
40 
41 	return s
42 }
43 
44 func NewCompactor() *Compactor {
45 	return &Compactor{ch: make(chan struct{}, 1)}
46 }
47 
48 func (c *Compactor) Start(ctx context.Context, stopper *Stopper) {
49 	c.ch <- struct{}{}
50 	stopper.RunWorker(ctx, func(ctx context.Context) {           // <-------
51 		for {
52 			select {
53 			case <-stopper.ShouldStop():
54 				return
55 			case <-c.ch:
56 			}
57 		}
58 	})
59 }
60 
61 
62 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/25456/cockroach25456_test.go#52
```go
41 ...
42 
43 
44 func NewReplica(store *Store) *Replica {
45 	return &Replica{store: store}
46 }
47 
48 type consistencyQueue struct{}
49 
50 func (q *consistencyQueue) process(repl *Replica) {
51 	<-repl.store.Stopper().ShouldQuiesce()
52 }           // <-------
53 
54 func newConsistencyQueue() *consistencyQueue {
55 	return &consistencyQueue{}
56 }
57 
58 type testContext struct {
59 	store *Store
60 	repl  *Replica
61 }
62 
63 
64 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/35931/cockroach35931_test.go#22
```go
11 ...
12 
13 	receiver RowReceiver
14 }
15 
16 type RowChannel struct {
17 	dataChan chan struct{}
18 }
19 
20 func (rc *RowChannel) Push() {
21 	rc.dataChan <- struct{}{}
22 }           // <-------
23 
24 func (rc *RowChannel) initWithBufSizeAndNumSenders(chanBufSize int) {
25 	rc.dataChan = make(chan struct{}, chanBufSize)
26 }
27 
28 type flowEntry struct {
29 	flow           *Flow
30 	inboundStreams map[int]*inboundStreamInfo
31 }
32 
33 
34 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/584/cockroach584_test.go#28
```go
17 ...
18 
19 		}
20 		g.mu.Unlock()
21 		break
22 	}
23 }
24 
25 func (g *Gossip) manage() {
26 	for {
27 		g.mu.Lock()
28 		if g.closed {           // <-------
29 			/// Missing g.mu.Unlock
30 			break
31 		}
32 		g.mu.Unlock()
33 		break
34 	}
35 }
36 func TestCockroach584(t *testing.T) {
37 	g := &Gossip{
38 		closed: true,
39 
40 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/9935/cockroach9935_test.go#38
```go
27 ...
28 
29 }
30 func (l *loggingT) createFile() error {
31 	if rand.Intn(8)%4 > 0 {
32 		return errors.New("")
33 	}
34 	return nil
35 }
36 func (l *loggingT) exit(err error) {
37 	l.mu.Lock()
38 	defer l.mu.Unlock()           // <-------
39 }
40 func TestCockroach9935(t *testing.T) {
41 	l := &loggingT{}
42 	go l.outputLogEntry()
43 }
44 
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/etcd/6708/etcd6708_test.go#50
```go
39 ...
40 
41 	switch c.selectionMode {
42 	case EndpointSelectionRandom:
43 	case EndpointSelectionPrioritizeLeader:
44 		c.getLeaderEndpoint()
45 	}
46 }
47 
48 func (c *httpClusterClient) Do(ctx context.Context) {
49 	c.RLock() // block here
50 	c.RUnlock()           // <-------
51 }
52 
53 func (c *httpClusterClient) Sync(ctx context.Context) {
54 	c.Lock()
55 	defer c.Unlock()
56 
57 	c.SetEndpoints()
58 }
59 
60 type httpMembersAPI struct {
61 
62 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/etcd/6857/etcd6857_test.go#25
```go
14 ...
15 
16 type node struct {
17 	status chan chan Status
18 	stop   chan struct{}
19 	done   chan struct{}
20 }
21 
22 func (n *node) Status() Status {
23 	c := make(chan Status)
24 	n.status <- c
25 	return <-c           // <-------
26 }
27 
28 func (n *node) run() {
29 	for {
30 		select {
31 		case c := <-n.status:
32 			c <- Status{}
33 		case <-n.stop:
34 			close(n.done)
35 			return
36 
37 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/etcd/7492/etcd7492_test.go#70
```go
59 ...
60 
61 		case waitCh := <-tm.stopCh:
62 			waitCh <- struct{}{}
63 			return
64 		}
65 	}
66 }
67 
68 func (tm *simpleTokenTTLKeeper) addSimpleToken() {
69 	tm.addSimpleTokenCh <- struct{}{}
70 }           // <-------
71 
72 func (tm *simpleTokenTTLKeeper) stop() {
73 	waitCh := make(chan struct{})
74 	tm.stopCh <- waitCh
75 	<-waitCh
76 	close(tm.stopCh)
77 }
78 
79 type tokenSimple struct {
80 	simpleTokenKeeper *simpleTokenTTLKeeper
81 
82 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/etcd/7492/etcd7492_test.go#162


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/etcd/7902/etcd7902_test.go#64
```go
53 ...
54 
55 	var mu sync.Mutex
56 	var wg sync.WaitGroup
57 	wg.Add(len(rcs))
58 	for i := range rcs {
59 		go func(rc *roundClient) { // G2,G3
60 			defer wg.Done()
61 			for rc.progress < rounds || rounds <= 0 {
62 				rc.acquire()
63 				mu.Lock()
64 				rc.validate()           // <-------
65 				mu.Unlock()
66 				rc.progress++
67 				mu.Lock() // Leader is blocking here
68 				rc.release()
69 				mu.Unlock()
70 			}
71 		}(&rcs[i])
72 	}
73 	wg.Wait()
74 }
75 
76 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/etcd/7902/etcd7902_test.go#50
```go
39 ...
40 
41 		rcs[i].validate = func() {
42 			setRcNextc()
43 		}
44 		rcs[i].release = func() {
45 			if i == 0 { // Assume the first roundClient is the leader
46 				close(nextc)
47 				nextc = make(chan bool)
48 			}
49 			<-rcNextc // Followers is blocking here
50 		}           // <-------
51 	}
52 	doRounds(rcs, 100)
53 }
54 func doRounds(rcs []roundClient, rounds int) {
55 	var mu sync.Mutex
56 	var wg sync.WaitGroup
57 	wg.Add(len(rcs))
58 	for i := range rcs {
59 		go func(rc *roundClient) { // G2,G3
60 			defer wg.Done()
61 
62 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/etcd/7902/etcd7902_test.go#74
```go
63 ...
64 
65 				mu.Unlock()
66 				rc.progress++
67 				mu.Lock() // Leader is blocking here
68 				rc.release()
69 				mu.Unlock()
70 			}
71 		}(&rcs[i])
72 	}
73 	wg.Wait()
74 }           // <-------
75 
76 ///
77 /// G1						G2 (leader)					G3 (follower)
78 /// runElectionFunc()
79 /// doRounds()
80 /// wg.Wait()
81 /// 						...
82 /// 						mu.Lock()
83 /// 						rc.validate()
84 /// 						rcNextc = nextc
85 
86 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/grpc/1275/grpc1275_test.go#41
```go
30 ...
31 
32 }
33 
34 type recvBufferReader struct {
35 	recv *recvBuffer
36 }
37 
38 func (r *recvBufferReader) Read(p []byte) (int, error) {
39 	select {
40 	case <-r.recv.get(): // G2 block here
41 	}           // <-------
42 	return 0, nil
43 }
44 
45 type Stream struct {
46 	trReader io.Reader
47 }
48 
49 func (s *Stream) Read(p []byte) (int, error) {
50 	return io.ReadFull(s.trReader, p)
51 }
52 
53 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/grpc/1353/grpc1353_test.go#167


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/grpc/1424/grpc1424_test.go#87
```go
76 ...
77 
78 	cc := NewClientConn()
79 	waitC := make(chan error, 1)
80 	go func() { // G2
81 		defer close(waitC)
82 		ch := cc.dopts.balancer.Notify()
83 		if ch != nil {
84 			doneChan := make(chan bool)
85 			go cc.lbWatcher(doneChan) // G3
86 			<-doneChan                /// Block here
87 		}           // <-------
88 	}()
89 	/// close addrCh
90 	close(cc.dopts.balancer.(*roundRobin).addrCh)
91 }
92 
93 ///
94 /// G1                      G2                          G3
95 /// DialContext()
96 ///                         cc.dopts.balancer.Notify()
97 ///                                                     cc.lbWatcher()
98 
99 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/grpc/3017/grpc3017_test.go#95
```go
84 ...
85 
86 
87 	sc := ccc.NewSubConn([]Address{Address(1)})
88 	go func() {
89 		for i := 0; i < 1000; i++ {
90 			ccc.RemoveSubConn(sc)
91 			sc = ccc.NewSubConn([]Address{Address(1)})
92 		}
93 		close(done)
94 	}()
95 	<-done           // <-------
96 }
97 
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/grpc/3017/grpc3017_test.go#43
```go
32 ...
33 
34 		delete(ccc.subConnCache, addrWithoutMD)
35 		return entry.sc
36 	}
37 	scNew := SubConn(1)
38 	ccc.subConnToAddr[scNew] = addrWithoutMD
39 	return scNew
40 }
41 
42 func (ccc *lbCacheClientConn) RemoveSubConn(sc SubConn) {
43 	ccc.mu.Lock()           // <-------
44 	defer ccc.mu.Unlock()
45 	addr, ok := ccc.subConnToAddr[sc]
46 	if !ok {
47 		return
48 	}
49 
50 	if entry, ok := ccc.subConnCache[addr]; ok {
51 		if entry.sc != sc {
52 			delete(ccc.subConnToAddr, sc)
53 		}
54 
55 ...
```


## Replay
**Replaying was not run**.

