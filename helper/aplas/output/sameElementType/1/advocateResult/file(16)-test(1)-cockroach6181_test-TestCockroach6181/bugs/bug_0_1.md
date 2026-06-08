# Bug: A07 - Actual Non-Cyclic Blocking Bug

During the execution, a blocking bug was detected.
This means, there is a routine that is blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach6181
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/6181/cockroach6181_test.go
- Trace: advocateTrace_17

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/1055/cockroach1055_test.go#83
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/1055/cockroach1055_test.go#49
```go
38 ...
39 
40 }
41 
42 func (s *Stopper) Stop() {
43 	s.mu.Lock()
44 	defer s.mu.Unlock()
45 	atomic.StoreInt32(&s.draining, 1)
46 	s.drain.Wait()
47 	close(s.stopper)
48 	s.stop.Wait()
49 }           // <-------
50 
51 func (s *Stopper) StartTask() bool {
52 	if atomic.LoadInt32(&s.draining) == 0 {
53 		s.mu.Lock()
54 		defer s.mu.Unlock()
55 		s.drain.Add(1)
56 		return true
57 	}
58 	return false
59 }
60 
61 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/1055/cockroach1055_test.go#99


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/1055/cockroach1055_test.go#54
```go
43 ...
44 
45 	atomic.StoreInt32(&s.draining, 1)
46 	s.drain.Wait()
47 	close(s.stopper)
48 	s.stop.Wait()
49 }
50 
51 func (s *Stopper) StartTask() bool {
52 	if atomic.LoadInt32(&s.draining) == 0 {
53 		s.mu.Lock()
54 		defer s.mu.Unlock()           // <-------
55 		s.drain.Add(1)
56 		return true
57 	}
58 	return false
59 }
60 
61 func NewStopper() *Stopper {
62 	return &Stopper{
63 		stopper: make(chan struct{}),
64 	}
65 
66 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/13197/cockroach13197_test.go#36
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/13755/cockroach13755_test.go#30
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/18101/cockroach18101_test.go#41
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/24808/cockroach24808_test.go#50
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/25456/cockroach25456_test.go#52
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/35931/cockroach35931_test.go#22
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


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/1/blocking/cockroach/584/cockroach584_test.go#28
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


## Replay
**Replaying was not run**.

