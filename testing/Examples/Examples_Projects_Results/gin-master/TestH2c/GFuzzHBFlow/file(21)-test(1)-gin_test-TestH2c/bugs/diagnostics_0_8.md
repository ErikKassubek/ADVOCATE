# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestH2c
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/server.go:1182
```go
1171 ...
1172 
1173 	err := sc.writeFrameFromHandler(FrameWriteRequest{
1174 		write:  writeArg,
1175 		stream: stream,
1176 		done:   ch,
1177 	})
1178 	if err != nil {
1179 		return err
1180 	}
1181 	var frameWriteDone bool // the frame write is done (successfully or not)
1182 	select {           // <-------
1183 	case err = <-ch:
1184 		frameWriteDone = true
1185 	case <-sc.doneServing:
1186 		return errClientDisconnected
1187 	case <-stream.cw:
1188 		// If both ch and stream.cw were ready (as might
1189 		// happen on the final Write after an http.Handler
1190 		// ends), prefer the write result. Otherwise this
1191 		// might just be us successfully closing the stream.
1192 		// The writeFrameAsync and serve goroutines guarantee
1193 
1194 ...
```


###  Channel: Close
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/http2.go:247
```go
236 ...
237 
238 // It exists because so a closeWaiter value can be placed inside a
239 // larger struct and have the Mutex and Cond's memory in the same
240 // allocation.
241 func (cw *closeWaiter) Init() {
242 	*cw = make(chan struct{})
243 }
244 
245 // Close marks the closeWaiter as closed and unblocks any waiters.
246 func (cw closeWaiter) Close() {
247 	close(cw)           // <-------
248 }
249 
250 // Wait waits for the closeWaiter to become closed.
251 func (cw closeWaiter) Wait() {
252 	<-cw
253 }
254 
255 // bufferedWriter is a buffered writer that writes to w.
256 // Its buffered writer is lazily allocated as needed, to minimize
257 // idle memory usage with many connections.
258 
259 ...
```


## Replay
**Replaying was not run**.

