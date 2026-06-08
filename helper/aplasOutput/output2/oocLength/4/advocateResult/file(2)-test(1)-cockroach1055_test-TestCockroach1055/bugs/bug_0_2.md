# Bug: A07 - Actual Non-Cyclic Blocking Bug

During the execution, a blocking bug was detected.
This means, there is a routine that is blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach1055
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/1055/cockroach1055_test.go
- Trace: advocateTrace_3

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/1055/cockroach1055_test.go#78
```go
67 ...
68 
69 	for i := 0; i < 3; i++ {
70 		stoppers = append(stoppers, NewStopper())
71 	}
72 
73 	for i := range stoppers {
74 		s := stoppers[i]
75 		s.AddWorker()
76 		go func() {
77 			s.StartTask()
78 			<-s.ShouldStop()           // <-------
79 			s.SetStopped()
80 		}()
81 	}
82 
83 	done := make(chan struct{})
84 	go func() {
85 		for _, s := range stoppers {
86 			s.Quiesce()
87 		}
88 		for _, s := range stoppers {
89 
90 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/1055/cockroach1055_test.go#46
```go
35 ...
36 
37 	s.draining = 1
38 	s.drain.Wait()
39 	s.draining = 0
40 }
41 
42 func (s *Stopper) Stop() {
43 	s.mu.Lock()
44 	defer s.mu.Unlock()
45 	atomic.StoreInt32(&s.draining, 1)
46 	s.drain.Wait()           // <-------
47 	close(s.stopper)
48 	s.stop.Wait()
49 }
50 
51 func (s *Stopper) StartTask() bool {
52 	if atomic.LoadInt32(&s.draining) == 0 {
53 		s.mu.Lock()
54 		defer s.mu.Unlock()
55 		s.drain.Add(1)
56 		return true
57 
58 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/1055/cockroach1055_test.go#94
```go
83 ...
84 
85 		for _, s := range stoppers {
86 			s.Quiesce()
87 		}
88 		for _, s := range stoppers {
89 			s.Stop()
90 		}
91 		close(done)
92 	}()
93 
94 	<-done           // <-------
95 }
96 
```


## Replay
**Replaying was not run**.

