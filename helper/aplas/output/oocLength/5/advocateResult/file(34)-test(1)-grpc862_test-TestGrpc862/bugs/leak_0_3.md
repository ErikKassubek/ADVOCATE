# Leak: L07 - Block on select

The analyzer detected a Block on a select.
A Block on a select is a situation, where a select is still blocking at the end of the program.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestGrpc862
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/grpc/862/grpc862_test.go
- Trace: advocateTrace_35

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Select:
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/grpc/862/grpc862_test.go#57
```go
46 ...
47 
48 }
49 
50 func (ac *addrConn) resetTransport() {
51 	for retries := 1; ; retries++ {
52 		sleepTime := 2 * time.Nanosecond * time.Duration(retries)
53 		timeout := 10 * time.Nanosecond
54 		_, cancel := context.WithTimeout(ac.ctx, timeout)
55 		connectTime := time.Now()
56 		cancel()
57 		select { // Block here           // <-------
58 		case <-time.After(sleepTime - time.Since(connectTime)):
59 		case <-ac.ctx.Done():
60 			return
61 		}
62 	}
63 }
64 
65 func (ac *addrConn) tearDown() {
66 	ac.cancel()
67 }
68 
69 ...
```


## Replay
**Replaying was not run**.

