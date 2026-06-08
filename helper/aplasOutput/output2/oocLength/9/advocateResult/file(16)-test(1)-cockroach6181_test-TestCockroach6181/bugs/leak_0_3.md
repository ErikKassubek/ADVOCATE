# Leak: L08 - Leak on Sync.Mutex

The analyzer detected a leak on a sync.Mutex.
A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.
A Leak could potentially resolve itself, if the program would run longer.
A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach6181
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/9/blocking/cockroach/6181/cockroach6181_test.go
- Trace: advocateTrace_17

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Mutex: RLock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/9/blocking/cockroach/6181/cockroach6181_test.go#37
```go
26 ...
27 
28 func (rdc *rangeDescriptorCache) LookupRangeDescriptor() {
29 	rdc.rangeCacheMu.RLock()
30 	fmt.Printf("lookup range descriptor: %s", rdc)
31 	rdc.rangeCacheMu.RUnlock()
32 	rdc.rangeCacheMu.Lock()
33 	rdc.rangeCacheMu.Unlock()
34 }
35 
36 func (rdc *rangeDescriptorCache) String() string {
37 	rdc.rangeCacheMu.RLock()           // <-------
38 	defer rdc.rangeCacheMu.RUnlock()
39 	return rdc.stringLocked()
40 }
41 
42 func (rdc *rangeDescriptorCache) stringLocked() string {
43 	return "something here"
44 }
45 
46 func doLookupWithToken(rc *rangeDescriptorCache) {
47 	rc.LookupRangeDescriptor()
48 
49 ...
```


###  Mutex: RLock
## Replay
**Replaying was not run**.

