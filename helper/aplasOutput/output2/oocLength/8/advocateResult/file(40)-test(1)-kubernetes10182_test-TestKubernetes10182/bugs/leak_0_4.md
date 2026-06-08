# Leak: L08 - Leak on Sync.Mutex

The analyzer detected a leak on a sync.Mutex.
A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.
A Leak could potentially resolve itself, if the program would run longer.
A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestKubernetes10182
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/kubernetes/10182/kubernetes10182_test.go
- Trace: advocateTrace_41

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Mutex: Lock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/kubernetes/10182/kubernetes10182_test.go#38
```go
27 ...
28 
29 	}()
30 }
31 
32 func (s *statusManager) syncBatch() {
33 	<-s.podStatusChannel
34 	s.DeletePodStatus()
35 }
36 
37 func (s *statusManager) DeletePodStatus() {
38 	s.podStatusesLock.Lock()           // <-------
39 	defer s.podStatusesLock.Unlock()
40 }
41 
42 func (s *statusManager) SetPodStatus() {
43 	s.podStatusesLock.Lock()
44 	defer s.podStatusesLock.Unlock()
45 	s.podStatusChannel <- true
46 }
47 
48 func NewStatusManager() *statusManager {
49 
50 ...
```


###  Mutex: Lock
## Replay
**Replaying was not run**.

