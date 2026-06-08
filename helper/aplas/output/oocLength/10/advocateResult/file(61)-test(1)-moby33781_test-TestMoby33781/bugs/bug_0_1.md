# Bug: P01 - Possible Send on Closed Channel

The analyzer detected a possible send on a closed channel.
Although the send on a closed channel did not occur during the recording, it is possible that it will occur, based on the happens before relation.
Such a send on a closed channel leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMoby33781
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/moby/33781/moby33781_test.go
- Trace: advocateTrace_62

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/moby/33781/moby33781_test.go#33
```go
22 ...
23 
24 	probeTimeout := 50 * time.Nanosecond
25 	for {
26 		select {
27 		case <-stop:
28 			return
29 		case <-time.After(probeInterval):
30 			results := make(chan bool)
31 			ctx, cancelProbe := context.WithTimeout(context.Background(), probeTimeout)
32 			go func() { // G3
33 				results <- true           // <-------
34 				close(results)
35 			}()
36 			select {
37 			case <-stop:
38 				// results should be drained here
39 				cancelProbe()
40 				return
41 			case <-results:
42 				cancelProbe()
43 			case <-ctx.Done():
44 
45 ...
```


###  Channel: Close
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/moby/33781/moby33781_test.go#34
```go
23 ...
24 
25 	for {
26 		select {
27 		case <-stop:
28 			return
29 		case <-time.After(probeInterval):
30 			results := make(chan bool)
31 			ctx, cancelProbe := context.WithTimeout(context.Background(), probeTimeout)
32 			go func() { // G3
33 				results <- true
34 				close(results)           // <-------
35 			}()
36 			select {
37 			case <-stop:
38 				// results should be drained here
39 				cancelProbe()
40 				return
41 			case <-results:
42 				cancelProbe()
43 			case <-ctx.Done():
44 				cancelProbe()
45 
46 ...
```


## Replay
**Replaying was not run**.

