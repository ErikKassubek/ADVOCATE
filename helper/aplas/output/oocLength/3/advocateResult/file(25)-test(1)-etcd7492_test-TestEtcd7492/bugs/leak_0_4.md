# Leak: L08 - Leak on Sync.Mutex

The analyzer detected a leak on a sync.Mutex.
A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.
A Leak could potentially resolve itself, if the program would run longer.
A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestEtcd7492
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/etcd/7492/etcd7492_test.go
- Trace: advocateTrace_26

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Mutex: Lock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/etcd/7492/etcd7492_test.go#95
```go
84 ...
85 
86 }
87 
88 func (t *tokenSimple) assignSimpleTokenToUser() {
89 	t.simpleTokensMu.Lock()
90 	t.simpleTokenKeeper.addSimpleToken()
91 	t.simpleTokensMu.Unlock()
92 }
93 func newDeleterFunc(t *tokenSimple) func(string) {
94 	return func(tk string) {
95 		t.simpleTokensMu.Lock()           // <-------
96 		defer t.simpleTokensMu.Unlock()
97 	}
98 }
99 
100 func (t *tokenSimple) enable() {
101 	t.simpleTokenKeeper = NewSimpleTokenTTLKeeper(newDeleterFunc(t))
102 }
103 
104 func (t *tokenSimple) disable() {
105 	if t.simpleTokenKeeper != nil {
106 
107 ...
```


###  Mutex: Lock
## Replay
**Replaying was not run**.

