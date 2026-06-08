# Bug: A01 - Actual Send on Closed Channel

During the execution of the program, a send on a closed channel occurred.
The occurrence of a send on closed leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestServing5865
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/nonblocking/serving/5865/serving5865_test.go
- Trace: advocateTrace_83

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/nonblocking/serving/5865/serving5865_test.go#26
```go
15 ...
16 
17 	revisionWatchersMux sync.RWMutex
18 }
19 
20 func newRevisionWatcher(destsCh chan struct{}) *revisionWatcher {
21 	return &revisionWatcher{destsCh: destsCh}
22 }
23 
24 func (rbm *revisionBackendsManager) endpointsUpdated() {
25 	rw := rbm.getOrCreateRevisionWatcher()
26 	rw.destsCh <- struct{}{}           // <-------
27 }
28 
29 func (rbm *revisionBackendsManager) getOrCreateRevisionWatcher() *revisionWatcher {
30 	rbm.revisionWatchersMux.Lock()
31 	defer rbm.revisionWatchersMux.Unlock()
32 
33 	destsCh := make(chan struct{})
34 	rw := newRevisionWatcher(destsCh)
35 	go rw.run()
36 
37 
38 ...
```


###  Channel: Close
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/nonblocking/serving/5865/serving5865_test.go#14
```go
3 ...
4 
5 	"testing"
6 )
7 
8 type revisionWatcher struct {
9 	destsCh chan struct{}
10 }
11 
12 func (rw *revisionWatcher) run() {
13 	defer close(rw.destsCh)
14 }           // <-------
15 
16 type revisionBackendsManager struct {
17 	revisionWatchersMux sync.RWMutex
18 }
19 
20 func newRevisionWatcher(destsCh chan struct{}) *revisionWatcher {
21 	return &revisionWatcher{destsCh: destsCh}
22 }
23 
24 func (rbm *revisionBackendsManager) endpointsUpdated() {
25 
26 ...
```


## Replay
**Replaying was not run**.

