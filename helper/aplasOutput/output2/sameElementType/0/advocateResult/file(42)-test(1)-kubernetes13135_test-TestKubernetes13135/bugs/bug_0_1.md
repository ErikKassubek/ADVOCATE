# Bug: P05 - Possible Cyclic Deadlock

The analysis detected a possible cyclic deadlock.
If this deadlock contains or influences the run of the main routine, this can result in the program getting stuck. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestKubernetes13135
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/0/blocking/kubernetes/13135/kubernetes13135_test.go
- Trace: advocateTrace_43

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/0/blocking/kubernetes/13135/kubernetes13135_test.go#112
```go
101 ...
102 
103 		w.onEvent()
104 	}
105 }
106 
107 func (w *WatchCache) Add(obj interface{}) {
108 	w.processEvent()
109 }
110 
111 func (w *WatchCache) Replace(obj interface{}) {
112 	w.Lock()           // <-------
113 	defer w.Unlock()
114 	if w.onReplace != nil {
115 		w.onReplace()
116 	}
117 }
118 
119 func NewCacher() *Cacher {
120 	watchCache := &WatchCache{}
121 	cacher := &Cacher{
122 		initialized: sync.WaitGroup{},
123 
124 ...
```


###  Mutex: Part of deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/0/blocking/kubernetes/13135/kubernetes13135_test.go#67
```go
56 ...
57 
58 type Cacher struct {
59 	sync.Mutex
60 	initialized sync.WaitGroup
61 	initOnce    sync.Once
62 	watchCache  *WatchCache
63 	reflector   *Reflector
64 }
65 
66 func (c *Cacher) processEvent() {
67 	c.Lock()           // <-------
68 	defer c.Unlock()
69 }
70 
71 func (c *Cacher) startCaching(stopChannel <-chan struct{}) {
72 	c.Lock()
73 	for {
74 		err := c.reflector.ListAndWatch(stopChannel)
75 		if err == nil {
76 			break
77 		}
78 
79 ...
```


## Replay
**Replaying was not run**.

