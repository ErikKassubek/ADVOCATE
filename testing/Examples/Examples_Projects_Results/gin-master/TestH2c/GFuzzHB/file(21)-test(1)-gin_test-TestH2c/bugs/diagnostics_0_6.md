# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestH2c
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/client_conn_pool.go:97
```go
86 ...
87 
88 			}
89 		}
90 		if !dialOnMiss {
91 			p.mu.Unlock()
92 			return nil, ErrNoCachedConn
93 		}
94 		traceGetConn(req, addr)
95 		call := p.getStartDialLocked(req.Context(), addr)
96 		p.mu.Unlock()
97 		<-call.done           // <-------
98 		if shouldRetryDial(call, req) {
99 			continue
100 		}
101 		cc, err := call.res, call.err
102 		if err != nil {
103 			return nil, err
104 		}
105 		if cc.ReserveNewRequest() {
106 			return cc, nil
107 		}
108 
109 ...
```


###  Channel: Close
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/client_conn_pool.go:150
```go
139 ...
140 
141 	c.res, c.err = c.p.t.dialClientConn(ctx, addr, singleUse)
142 
143 	c.p.mu.Lock()
144 	delete(c.p.dialing, addr)
145 	if c.err == nil {
146 		c.p.addConnLocked(addr, c.res)
147 	}
148 	c.p.mu.Unlock()
149 
150 	close(c.done)           // <-------
151 }
152 
153 // addConnIfNeeded makes a NewClientConn out of c if a connection for key doesn't
154 // already exist. It coalesces concurrent calls with the same key.
155 // This is used by the http1 Transport code when it creates a new connection. Because
156 // the http1 Transport doesn't de-dup TCP dials to outbound hosts (because it doesn't know
157 // the protocol), it can get into a situation where it has multiple TLS connections.
158 // This code decides which ones live or die.
159 // The return value used is whether c was used.
160 // c is never closed.
161 
162 ...
```


## Replay
**Replaying was not run**.

