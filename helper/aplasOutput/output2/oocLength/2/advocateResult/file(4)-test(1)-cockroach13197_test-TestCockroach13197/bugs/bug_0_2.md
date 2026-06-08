# Bug: A08 - Actual Cyclic Deadlock

During the execution, a deadlock was detected.
This means, there is are routine that are cyclicaly blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach13197
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/2/blocking/cockroach/13197/cockroach13197_test.go
- Trace: advocateTrace_5

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/2/blocking/cockroach/10790/cockroach10790_test.go#63
```go
52 ...
53 
54 func (r *Replica) beginCmds(ctx context.Context) {
55 	ctxDone := ctx.Done()
56 	for _, ch := range r.chans {
57 		select {
58 		case <-ch:
59 		case <-ctxDone:
60 			go func() {
61 				for _, ch := range r.chans {
62 					<-ch
63 				}           // <-------
64 			}()
65 		}
66 	}
67 }
68 
69 /// helper goroutine, not present in the real bug.
70 func (r *Replica) sendChans(ctx context.Context) {
71 	for _, ch := range r.chans {
72 		select {
73 		case ch <- true:
74 
75 ...
```


## Replay
**Replaying was not run**.

