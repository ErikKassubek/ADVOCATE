# Leak: L08 - Leak on Sync.Mutex

The analyzer detected a leak on a sync.Mutex.
A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.
A Leak could potentially resolve itself, if the program would run longer.
A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach584
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/584/cockroach584_test.go
- Trace: advocateTrace_16

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Mutex: Lock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/3710/cockroach3710_test.go#47
```go
36 ...
37 
38 	s.mu.RLock()
39 	defer s.mu.RUnlock()
40 }
41 
42 func (s *Store) processRaft() {
43 	go func() {
44 		for {
45 			var replicas []*Replica
46 			s.mu.Lock()
47 			for _, r := range s.replicas {           // <-------
48 				replicas = append(replicas, r)
49 			}
50 			s.mu.Unlock()
51 			break
52 		}
53 	}()
54 }
55 
56 type Replica struct {
57 	store *Store
58 
59 ...
```


###  Mutex: Lock
## Replay
**Replaying was not run**.

