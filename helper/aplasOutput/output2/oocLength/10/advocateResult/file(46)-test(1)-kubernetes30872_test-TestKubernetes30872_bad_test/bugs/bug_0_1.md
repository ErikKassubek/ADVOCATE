# Bug: P05 - Possible Cyclic Deadlock

The analysis detected a possible cyclic deadlock.
If this deadlock contains or influences the run of the main routine, this can result in the program getting stuck. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestKubernetes30872_bad_test
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/kubernetes/30872/kubernetes30872_test.go
- Trace: advocateTrace_47

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/kubernetes/30872/kubernetes30872_test.go#157
```go
146 ...
147 
148 		nc.reconcileNamespace()
149 	})
150 }
151 
152 type DeltaFIFO struct {
153 	lock sync.RWMutex
154 }
155 
156 func (f *DeltaFIFO) HasSynced() {
157 	f.lock.Lock()           // <-------
158 	defer f.lock.Unlock()
159 }
160 
161 func (f *DeltaFIFO) Pop(process PopProcessFunc) {
162 	f.lock.Lock()
163 	defer f.lock.Unlock()
164 	process()
165 }
166 
167 func NewFederatedInformer() FederatedInformer {
168 
169 ...
```


###  Mutex: Part of deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/kubernetes/30872/kubernetes30872_test.go#92
```go
81 ...
82 
83 }
84 
85 func (f *federatedInformerImpl) ClustersSynced() {
86 	f.Lock()
87 	defer f.Unlock()
88 	f.clusterInformer.controller.HasSynced()
89 }
90 
91 func (f *federatedInformerImpl) addCluster() {
92 	f.Lock()           // <-------
93 	defer f.Unlock()
94 }
95 
96 func (f *federatedInformerImpl) Start() {
97 	f.Lock()
98 	defer f.Unlock()
99 
100 	f.clusterInformer.stopChan = make(chan struct{})
101 	go f.clusterInformer.controller.Run(f.clusterInformer.stopChan)
102 }
103 
104 ...
```


## Replay
**Replaying was not run**.

