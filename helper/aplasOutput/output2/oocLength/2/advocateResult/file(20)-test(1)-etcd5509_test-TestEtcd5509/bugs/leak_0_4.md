# Leak: L08 - Leak on Sync.Mutex

The analyzer detected a leak on a sync.Mutex.
A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.
A Leak could potentially resolve itself, if the program would run longer.
A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestEtcd5509
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/2/blocking/etcd/5509/etcd5509_test.go
- Trace: advocateTrace_21

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Mutex: Lock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/2/blocking/etcd/10492/etcd10492_test.go#20
```go
9 ...
10 
11 
12 type lessor struct {
13 	mu                 sync.RWMutex
14 	cp                 Checkpointer
15 	checkpointInterval time.Duration
16 }
17 
18 func (le *lessor) Checkpoint() {
19 	le.mu.Lock() // block here
20 	defer le.mu.Unlock()           // <-------
21 }
22 
23 func (le *lessor) SetCheckpointer(cp Checkpointer) {
24 	le.mu.Lock()
25 	defer le.mu.Unlock()
26 
27 	le.cp = cp
28 }
29 
30 func (le *lessor) Renew() {
31 
32 ...
```


###  Mutex: Lock
## Replay
**Replaying was not run**.

