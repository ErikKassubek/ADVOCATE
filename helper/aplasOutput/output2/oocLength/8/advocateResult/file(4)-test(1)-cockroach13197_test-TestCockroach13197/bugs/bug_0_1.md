# Bug: A07 - Actual Non-Cyclic Blocking Bug

During the execution, a blocking bug was detected.
This means, there is a routine that is blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach13197
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/13197/cockroach13197_test.go
- Trace: advocateTrace_5

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/1055/cockroach1055_test.go#83
```go
72 ...
73 
74 		s := stoppers[i]
75 		s.AddWorker()
76 		go func() {
77 			s.StartTask()
78 			<-s.ShouldStop()
79 			s.SetStopped()
80 		}()
81 	}
82 
83 	done := make(chan struct{})           // <-------
84 	go func() {
85 		for _, s := range stoppers {
86 			s.Quiesce()
87 		}
88 		for _, s := range stoppers {
89 			s.Stop()
90 		}
91 		close(done)
92 	}()
93 
94 
95 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/1055/cockroach1055_test.go#47
```go
36 ...
37 
38 	s.drain.Wait()
39 	s.draining = 0
40 }
41 
42 func (s *Stopper) Stop() {
43 	s.mu.Lock()
44 	defer s.mu.Unlock()
45 	atomic.StoreInt32(&s.draining, 1)
46 	s.drain.Wait()
47 	close(s.stopper)           // <-------
48 	s.stop.Wait()
49 }
50 
51 func (s *Stopper) StartTask() bool {
52 	if atomic.LoadInt32(&s.draining) == 0 {
53 		s.mu.Lock()
54 		defer s.mu.Unlock()
55 		s.drain.Add(1)
56 		return true
57 	}
58 
59 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/1055/cockroach1055_test.go#99


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/blocking/cockroach/13197/cockroach13197_test.go#35
```go
24 ...
25 
26 	return tx
27 }
28 
29 type Tx struct {
30 	cancel context.CancelFunc
31 	ctx    context.Context
32 }
33 
34 func (tx *Tx) awaitDone() {
35 	<-tx.ctx.Done()           // <-------
36 }
37 
38 func (tx *Tx) Rollback() {
39 	tx.rollback()
40 }
41 
42 func (tx *Tx) rollback() {
43 	tx.close()
44 }
45 
46 
47 ...
```


## Replay
**Replaying was not run**.

