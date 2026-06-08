# Bug: P05 - Possible Cyclic Deadlock

The analysis detected a possible cyclic deadlock.
If this deadlock contains or influences the run of the main routine, this can result in the program getting stuck. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestHugo3251
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/9/blocking/hugo/3251/hugo3251_test.go
- Trace: advocateTrace_36

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/9/blocking/hugo/3251/hugo3251_test.go#29
```go
18 ...
19 
20 	l.Lock()
21 	if _, ok := l.m[url]; !ok {
22 		l.m[url] = &sync.Mutex{}
23 	}
24 	l.m[url].Lock()
25 	l.Unlock()
26 }
27 
28 func (l *remoteLock) URLUnlock(url string) {
29 	l.RLock()           // <-------
30 	defer l.RUnlock()
31 	if um, ok := l.m[url]; ok {
32 		um.Unlock()
33 	}
34 }
35 
36 func resGetRemote(url string) error {
37 	remoteURLLock.URLLock(url)
38 	defer func() { remoteURLLock.URLUnlock(url) }()
39 
40 
41 ...
```


###  Mutex: Part of deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/9/blocking/hugo/3251/hugo3251_test.go#24
```go
13 ...
14 
15 	sync.RWMutex
16 	m map[string]*sync.Mutex
17 }
18 
19 func (l *remoteLock) URLLock(url string) {
20 	l.Lock()
21 	if _, ok := l.m[url]; !ok {
22 		l.m[url] = &sync.Mutex{}
23 	}
24 	l.m[url].Lock()           // <-------
25 	l.Unlock()
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
36 ...
```


## Replay
**Replaying was not run**.

