# Bug: P05 - Possible Cyclic Deadlock

The analysis detected a possible cyclic deadlock.
If this deadlock contains or influences the run of the main routine, this can result in the program getting stuck. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach10214
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/10214/cockroach10214_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/10214/cockroach10214_test.go#51
```go
40 ...
41 
42 }
43 
44 type Replica struct {
45 	raftMu sync.Mutex
46 	mu     sync.Mutex
47 	store  *Store
48 }
49 
50 func (r *Replica) reportUnreachable() {
51 	r.raftMu.Lock() // LockB acquire           // <-------
52 	//+time.Sleep(time.Nanosecond)
53 	defer r.raftMu.Unlock()
54 	// LockB release
55 }
56 
57 func (r *Replica) tick() {
58 	r.raftMu.Lock() // LockB acquire
59 	defer r.raftMu.Unlock()
60 	r.tickRaftMuLocked()
61 	// LockB release
62 
63 ...
```


###  Mutex: Part of deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/cockroach/10214/cockroach10214_test.go#83
```go
72 ...
73 
74 			return true
75 		}
76 	}
77 	return false
78 }
79 func (r *Replica) maybeCoalesceHeartbeat() bool {
80 	msgtype := uintptr(unsafe.Pointer(r)) % 3
81 	switch msgtype {
82 	case 0, 1, 2:
83 		r.store.coalescedMu.Lock() // LockA acquire           // <-------
84 	default:
85 		return false
86 	}
87 	r.store.coalescedMu.Unlock() // LockA release
88 	return true
89 }
90 
91 func TestCockroach10214(t *testing.T) {
92 	store := &Store{}
93 	responses := &store.coalescedMu.heartbeatResponses
94 
95 ...
```


## Replay
**Replaying was not run**.

