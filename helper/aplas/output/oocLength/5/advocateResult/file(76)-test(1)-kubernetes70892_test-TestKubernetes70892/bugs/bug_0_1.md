# Bug: P01 - Possible Send on Closed Channel

The analyzer detected a possible send on a closed channel.
Although the send on a closed channel did not occur during the recording, it is possible that it will occur, based on the happens before relation.
Such a send on a closed channel leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestKubernetes70892
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/nonblocking/kubernetes/70892/kubernetes70892_test.go
- Trace: advocateTrace_77

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/nonblocking/kubernetes/70892/kubernetes70892_test.go#21
```go
10 ...
11 
12 
13 func ParallelizeUntil(ctx context.Context, workers, pieces int, doWorkPiece DoWorkPieceFunc) {
14 	var stop <-chan struct{}
15 	if ctx != nil {
16 		stop = ctx.Done()
17 	}
18 
19 	toProcess := make(chan int, pieces)
20 	for i := 0; i < pieces; i++ {
21 		toProcess <- i           // <-------
22 	}
23 	close(toProcess)
24 
25 	if pieces < workers {
26 		workers = pieces
27 	}
28 
29 	wg := sync.WaitGroup{}
30 	wg.Add(workers)
31 	for i := 0; i < workers; i++ {
32 
33 ...
```


###  Channel: Close
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/nonblocking/kubernetes/70892/kubernetes70892_test.go#23
```go
12 ...
13 
14 	var stop <-chan struct{}
15 	if ctx != nil {
16 		stop = ctx.Done()
17 	}
18 
19 	toProcess := make(chan int, pieces)
20 	for i := 0; i < pieces; i++ {
21 		toProcess <- i
22 	}
23 	close(toProcess)           // <-------
24 
25 	if pieces < workers {
26 		workers = pieces
27 	}
28 
29 	wg := sync.WaitGroup{}
30 	wg.Add(workers)
31 	for i := 0; i < workers; i++ {
32 		go func() {
33 			defer wg.Done()
34 
35 ...
```


## Replay
**Replaying was not run**.

