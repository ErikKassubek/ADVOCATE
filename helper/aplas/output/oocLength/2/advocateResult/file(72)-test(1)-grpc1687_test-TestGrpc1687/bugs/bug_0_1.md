# Bug: A01 - Actual Send on Closed Channel

During the execution of the program, a send on a closed channel occurred.
The occurrence of a send on closed leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestGrpc1687
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/2/nonblocking/grpc/1687/grpc1687_test.go
- Trace: advocateTrace_73

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/2/nonblocking/grpc/1687/grpc1687_test.go#28
```go
17 ...
18 
19 	closedCh chan struct{}
20 	writes   chan func()
21 }
22 
23 func (ht *serverHandlerTransport) do(fn func()) {
24 	select {
25 	case <-ht.closedCh:
26 		return
27 	default:
28 		select {           // <-------
29 		case ht.writes <- fn:
30 			return
31 		case <-ht.closedCh:
32 			return
33 		}
34 	}
35 }
36 
37 func (ht *serverHandlerTransport) WriteStatus() {
38 	ht.do(func() {})
39 
40 ...
```


## Replay
**Replaying was not run**.

