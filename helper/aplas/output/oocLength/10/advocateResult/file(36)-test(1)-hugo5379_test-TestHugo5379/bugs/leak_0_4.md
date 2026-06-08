# Leak: L08 - Leak on Sync.Mutex

The analyzer detected a leak on a sync.Mutex.
A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.
A Leak could potentially resolve itself, if the program would run longer.
A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestHugo5379
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/hugo/5379/hugo5379_test.go
- Trace: advocateTrace_37

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Mutex: Lock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/hugo/3251/hugo3251_test.go#21
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


###  Mutex: Lock
## Replay
**Replaying was not run**.

