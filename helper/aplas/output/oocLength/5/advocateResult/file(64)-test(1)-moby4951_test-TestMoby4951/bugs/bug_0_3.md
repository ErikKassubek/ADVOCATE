# Bug: A07 - Actual Non-Cyclic Blocking Bug

During the execution, a blocking bug was detected.
This means, there is a routine that is blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMoby4951
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/4951/moby4951_test.go
- Trace: advocateTrace_65

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/1055/cockroach1055_test.go#83
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/1055/cockroach1055_test.go#47
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/1055/cockroach1055_test.go#99


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/13197/cockroach13197_test.go#36
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/13755/cockroach13755_test.go#30
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/18101/cockroach18101_test.go#41
```go
30 ...
31 
32 			return readyForImportSpan
33 		}
34 	}
35 	return true
36 }
37 
38 func splitAndScatter(ctx context.Context, readyForImportCh chan bool) {
39 	for i := 0; i < chanSize+2; i++ {
40 		readyForImportCh <- (false || i != 0)
41 	}           // <-------
42 }
43 
44 ///
45 /// G1					G2					helper goroutine
46 /// restore()
47 /// 					splitAndScatter()
48 /// <-readyForImportCh
49 /// 					readyForImportCh<-
50 /// ...					...
51 /// 										cancel()
52 
53 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/24808/cockroach24808_test.go#50
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/25456/cockroach25456_test.go#52
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/35931/cockroach35931_test.go#22
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/584/cockroach584_test.go#28
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/6181/cockroach6181_test.go#62
```go
51 ...
52 
53 		var wg sync.WaitGroup
54 		for i := 0; i < 3; i++ {
55 			wg.Add(1)
56 			go func() { // G2,G3,...
57 				doLookupWithToken(db.cache)
58 				wg.Done()
59 			}()
60 		}
61 		wg.Wait()
62 	}           // <-------
63 	pauseLookupResumeAndAssert()
64 }
65 
66 /// G1 									G2							G3					...
67 /// testRangeCacheCoalescedRquests()
68 /// initTestDescriptorDB()
69 /// pauseLookupResumeAndAssert()
70 /// return
71 /// 									doLookupWithToken()
72 ///																 	doLookupWithToken()
73 
74 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/cockroach/9935/cockroach9935_test.go#38
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/6708/etcd6708_test.go#50
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/6857/etcd6857_test.go#25
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/7902/etcd7902_test.go#74
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/7902/etcd7902_test.go#64
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/7902/etcd7902_test.go#50
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/grpc/1275/grpc1275_test.go#41
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/grpc/1353/grpc1353_test.go#167


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/grpc/1424/grpc1424_test.go#87
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/grpc/3017/grpc3017_test.go#100


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/grpc/660/grpc660_test.go#27
```go
16 ...
17 
18 	stop chan bool
19 }
20 
21 func (bc *benchmarkClient) doCloseLoopUnary() {
22 	for {
23 		done := make(chan bool)
24 		go func() { // G2
25 			if rand.Intn(10) > 7 {
26 				done <- false
27 				return           // <-------
28 			}
29 			done <- true
30 		}()
31 		select {
32 		case <-bc.stop:
33 			return
34 		case <-done:
35 		}
36 	}
37 }
38 
39 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/hugo/3251/hugo3251_test.go#30
```go
19 ...
20 
21 	if _, ok := l.m[url]; !ok {
22 		l.m[url] = &sync.Mutex{}
23 	}
24 	l.m[url].Lock()
25 	l.Unlock()
26 }
27 
28 func (l *remoteLock) URLUnlock(url string) {
29 	l.RLock()
30 	defer l.RUnlock()           // <-------
31 	if um, ok := l.m[url]; ok {
32 		um.Unlock()
33 	}
34 }
35 
36 func resGetRemote(url string) error {
37 	remoteURLLock.URLLock(url)
38 	defer func() { remoteURLLock.URLUnlock(url) }()
39 
40 	return nil
41 
42 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/hugo/3251/hugo3251_test.go#65


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/hugo/3251/hugo3251_test.go#25
```go
14 ...
15 
16 	m map[string]*sync.Mutex
17 }
18 
19 func (l *remoteLock) URLLock(url string) {
20 	l.Lock()
21 	if _, ok := l.m[url]; !ok {
22 		l.m[url] = &sync.Mutex{}
23 	}
24 	l.m[url].Lock()
25 	l.Unlock()           // <-------
26 }
27 
28 func (l *remoteLock) URLUnlock(url string) {
29 	l.RLock()
30 	defer l.RUnlock()
31 	if um, ok := l.m[url]; ok {
32 		um.Unlock()
33 	}
34 }
35 
36 
37 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/hugo/5379/hugo5379_test.go#184
```go
173 ...
174 
175 	numWorkers := 2
176 	wg := &sync.WaitGroup{}
177 
178 	for i := 0; i < numWorkers; i++ {
179 		wg.Add(1)
180 		go pageRenderer(s, wg)
181 	}
182 
183 	wg.Wait()
184 }           // <-------
185 
186 type sitesBuilder struct {
187 	H *HugoSites
188 }
189 
190 func (s *sitesBuilder) Build() *sitesBuilder {
191 	return s.build()
192 }
193 
194 func (s *sitesBuilder) build() *sitesBuilder {
195 
196 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/hugo/5379/hugo5379_test.go#67
```go
56 ...
57 
58 func (p *Page) initContentPlainAndMeta() {
59 	p.initContent()
60 	p.initPlain(true)
61 }
62 
63 func (p *Page) initPlain(lock bool) {
64 	p.plainInit.Do(func() {
65 		if lock {
66 			p.contentInitMu.Lock() /// Double locking here.
67 			defer p.contentInitMu.Unlock()           // <-------
68 		}
69 	})
70 }
71 
72 func (p *Page) withoutContent() *PageWithoutContent {
73 	p.pageInit.withoutContentInit.Do(func() {
74 		p.pageWithoutContent = &PageWithoutContent{Page: p}
75 	})
76 	return p.pageWithoutContent
77 }
78 
79 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/kubernetes/10182/kubernetes10182_test.go#46
```go
35 ...
36 
37 func (s *statusManager) DeletePodStatus() {
38 	s.podStatusesLock.Lock()
39 	defer s.podStatusesLock.Unlock()
40 }
41 
42 func (s *statusManager) SetPodStatus() {
43 	s.podStatusesLock.Lock()
44 	defer s.podStatusesLock.Unlock()
45 	s.podStatusChannel <- true
46 }           // <-------
47 
48 func NewStatusManager() *statusManager {
49 	return &statusManager{
50 		podStatusChannel: make(chan bool),
51 	}
52 }
53 
54 /// G1 						G2							G3
55 /// s.Start()
56 /// s.syncBatch()
57 
58 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/kubernetes/25331/kubernetes25331_test.go#39
```go
28 ...
29 
30 	wc.cancel()
31 }
32 
33 func (wc *watchChan) run() {
34 	select {
35 	case err := <-wc.errChan:
36 		errResult := len(err.Error()) != 0
37 		wc.cancel() // Removed in fix
38 		wc.resultChan <- errResult
39 	case <-wc.ctx.Done():           // <-------
40 	}
41 }
42 
43 func NewWatchChan() *watchChan {
44 	ctx, cancel := context.WithCancel(context.Background())
45 	return &watchChan{
46 		ctx:        ctx,
47 		cancel:     cancel,
48 		resultChan: make(chan bool),
49 		errChan:    make(chan error),
50 
51 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/kubernetes/38669/kubernetes38669_test.go#34
```go
23 ...
24 
25 		_, ok := <-c.input
26 		if !ok {
27 			return
28 		}
29 	}
30 }
31 
32 func (c *cacheWatcher) sendWatchCacheEvent(event *watchCacheEvent) {
33 	c.result <- Event(*event)
34 }           // <-------
35 
36 func (c *cacheWatcher) Stop() {
37 	c.stop()
38 }
39 
40 func (c *cacheWatcher) stop() {
41 	c.Lock()
42 	defer c.Unlock()
43 	if !c.stopped {
44 		c.stopped = true
45 
46 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/kubernetes/5316/kubernetes5316_test.go#30
```go
19 ...
20 
21 
22 func finishRequest(timeout time.Duration, fn func() error) {
23 	ch := make(chan bool)     // FIX: ch := make(chan bool, 1)
24 	errCh := make(chan error) // FIX: errCh := make(chan error, 1)
25 	go func() {               // G2
26 		if err := fn(); err != nil {
27 			errCh <- err
28 		} else {
29 			ch <- true
30 		}           // <-------
31 	}()
32 
33 	select {
34 	case <-ch:
35 	case <-errCh:
36 	case <-time.After(timeout):
37 	}
38 }
39 
40 ///
41 
42 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/kubernetes/6632/kubernetes6632_test.go#52
```go
41 ...
42 
43 }
44 
45 func (i *idleAwareFramer) WriteFrame() {
46 	i.writeLock.Lock()
47 	defer i.writeLock.Unlock()
48 	if i.resetChan == nil {
49 		return
50 	}
51 	i.resetChan <- true
52 }           // <-------
53 
54 func NewIdleAwareFramer() *idleAwareFramer {
55 	return &idleAwareFramer{
56 		resetChan: make(chan bool),
57 		conn: &Connection{
58 			closeChan: make(chan bool),
59 		},
60 	}
61 }
62 
63 
64 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/kubernetes/6632/kubernetes6632_test.go#37
```go
26 ...
27 
28 }
29 
30 func (i *idleAwareFramer) monitor() {
31 	var resetChan = i.resetChan
32 Loop:
33 	for {
34 		select {
35 		case <-i.conn.closeChan:
36 			i.writeLock.Lock()
37 			close(resetChan)           // <-------
38 			i.resetChan = nil
39 			i.writeLock.Unlock()
40 			break Loop
41 		}
42 	}
43 }
44 
45 func (i *idleAwareFramer) WriteFrame() {
46 	i.writeLock.Lock()
47 	defer i.writeLock.Unlock()
48 
49 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/kubernetes/70277/kubernetes70277_test.go#84


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/17176/moby17176_test.go#51
```go
40 ...
41 
42 
43 func testDevmapperLockReleasedDeviceDeletion() {
44 	ds := &DeviceSet{
45 		nrDeletedDevices: 0,
46 	}
47 	ds.cleanupDeletedDevices()
48 	doneChan := make(chan bool)
49 	go func() {
50 		ds.Lock()
51 		defer ds.Unlock()           // <-------
52 		doneChan <- true
53 	}()
54 
55 	select {
56 	case <-time.After(time.Millisecond):
57 	case <-doneChan:
58 	}
59 }
60 func TestMoby17176(t *testing.T) {
61 	testDevmapperLockReleasedDeviceDeletion()
62 
63 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/21233/moby21233_test.go#134
```go
123 ...
124 
125 		xrefs[i], watchers[i] = tm.Transfer(ChanOutput(progressChan)) /// Chan producer
126 	}
127 
128 	for i := range xrefs {
129 		xrefs[i].Release(watchers[i]) /// Block here
130 	}
131 
132 	close(progressChan)
133 	<-progressDone
134 }           // <-------
135 
136 ///
137 /// G1 						G2					G3
138 /// testTransfer()
139 /// tm.Transfer()
140 /// t.Watch()
141 /// 						WriteProgress()
142 /// 						ProgressChan<-
143 /// 											<-progressChan
144 /// 						...					...
145 
146 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/25384/moby25384_test.go#34
```go
23 ...
24 
25 
26 func (pm *Manager) init() {
27 	var group sync.WaitGroup
28 	group.Add(len(pm.plugins))
29 	for _, p := range pm.plugins {
30 		go func(p *plugin) {
31 			defer group.Done()
32 		}(p)
33 		group.Wait() // Block here
34 	}           // <-------
35 }
36 func TestMoby25384(t *testing.T) {
37 	p1 := &plugin{}
38 	p2 := &plugin{}
39 	pm := &Manager{
40 		plugins: []*plugin{p1, p2},
41 	}
42 	go pm.init()
43 }
44 
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/29733/moby29733_test.go#22
```go
11 ...
12 
13 type plugins struct {
14 	sync.Mutex
15 	plugins map[int]*Plugin
16 }
17 
18 func (p *Plugin) waitActive() {
19 	p.activateWait.L.Lock()
20 	for !p.activated {
21 		p.activateWait.Wait()
22 	}           // <-------
23 	p.activateWait.L.Unlock()
24 }
25 
26 type extpointHandlers struct {
27 	sync.RWMutex
28 	extpointHandlers map[int]struct{}
29 }
30 
31 var (
32 	storage  = plugins{plugins: make(map[int]*Plugin)}
33 
34 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/29733/moby29733_test.go#51
```go
40 ...
41 
42 }
43 
44 func testActive(p *Plugin) {
45 	done := make(chan struct{})
46 	go func() {
47 		p.waitActive()
48 		close(done)
49 	}()
50 	<-done
51 }           // <-------
52 
53 func TestMoby29733(t *testing.T) {
54 	p := &Plugin{activateWait: sync.NewCond(&sync.Mutex{})}
55 	storage.plugins[0] = p
56 
57 	testActive(p)
58 	Handle()
59 	testActive(p)
60 }
61 
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/30408/moby30408_test.go#39
```go
28 ...
29 
30 }
31 
32 func testActive(p *Plugin) {
33 	done := make(chan struct{})
34 	go func() {
35 		p.waitActive()
36 		close(done)
37 	}()
38 	<-done
39 }           // <-------
40 func TestMoby30408(t *testing.T) {
41 	p := &Plugin{activateWait: sync.NewCond(&sync.Mutex{})}
42 	p.activateErr = errors.New("some junk happened")
43 
44 	testActive(p)
45 }
46 
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/30408/moby30408_test.go#23
```go
12 ...
13 
14 	activateWait *sync.Cond
15 	activateErr  error
16 	Manifest     *Manifest
17 }
18 
19 func (p *Plugin) waitActive() error {
20 	p.activateWait.L.Lock()
21 	for !p.activated() {
22 		p.activateWait.Wait()
23 	}           // <-------
24 	p.activateWait.L.Unlock()
25 	return p.activateErr
26 }
27 
28 func (p *Plugin) activated() bool {
29 	return p.Manifest != nil
30 }
31 
32 func testActive(p *Plugin) {
33 	done := make(chan struct{})
34 
35 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/33293/moby33293_test.go#27
```go
16 ...
17 
18 		return errors.New("Error")
19 	}
20 	return nil
21 }
22 func containerWait() <-chan error {
23 	errC := make(chan error)
24 	err := MayReturnError()
25 	if err != nil {
26 		errC <- err /// Block here
27 		return errC           // <-------
28 	}
29 	return errC
30 }
31 
32 ///
33 /// G1
34 /// containerWait()
35 /// errC <- err
36 /// ---------G1 leak---------------
37 ///
38 
39 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/33781/moby33781_test.go#34
```go
23 ...
24 
25 	for {
26 		select {
27 		case <-stop:
28 			return
29 		case <-time.After(probeInterval):
30 			results := make(chan bool)
31 			ctx, cancelProbe := context.WithTimeout(context.Background(), probeTimeout)
32 			go func() { // G3
33 				results <- true
34 				close(results)           // <-------
35 			}()
36 			select {
37 			case <-stop:
38 				// results should be drained here
39 				cancelProbe()
40 				return
41 			case <-results:
42 				cancelProbe()
43 			case <-ctx.Done():
44 				cancelProbe()
45 
46 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/36114/moby36114_test.go#31
```go
20 ...
21 
22 
23 func (svm *serviceVM) hotAddVHDsAtStart() {
24 	svm.Lock()
25 	defer svm.Unlock()
26 	svm.hotRemoveVHDsAtStart()
27 }
28 
29 func (svm *serviceVM) hotRemoveVHDsAtStart() {
30 	svm.Lock() // Double lock here
31 	defer svm.Unlock()           // <-------
32 }
33 
34 func TestMoby36114(t *testing.T) {
35 	s := &serviceVM{}
36 	go s.hotAddVHDsAtStart()
37 }
38 
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/moby/4395/moby4395_test.go#23
```go
12 ...
13 
14 import (
15 	"errors"
16 	"testing"
17 )
18 
19 func Go(f func() error) chan error {
20 	ch := make(chan error)
21 	go func() {
22 		ch <- f() // G2
23 	}()           // <-------
24 	return ch
25 }
26 
27 ///
28 /// G1				G2
29 /// Go()
30 /// return ch
31 /// 				ch <- f()
32 /// ----------G2 leak-------------
33 ///
34 
35 ...
```


## Replay
**Replaying was not run**.

