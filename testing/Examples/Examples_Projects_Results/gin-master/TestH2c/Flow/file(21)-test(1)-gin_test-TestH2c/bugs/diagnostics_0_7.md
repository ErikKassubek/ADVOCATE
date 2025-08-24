# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestH2c
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/transport.go:2368
```go
2357 ...
2358 
2359 		})
2360 		return nil // return nil from process* funcs to keep conn alive
2361 	}
2362 	if res == nil {
2363 		// (nil, nil) special case. See handleResponse docs.
2364 		return nil
2365 	}
2366 	cs.resTrailer = &res.Trailer
2367 	cs.res = res
2368 	close(cs.respHeaderRecv)           // <-------
2369 	if f.StreamEnded() {
2370 		rl.endStream(cs)
2371 	}
2372 	return nil
2373 }
2374 
2375 // may return error types nil, or ConnectionError. Any other error value
2376 // is a StreamError of type ErrCodeProtocol. The returned error in that case
2377 // is the detail.
2378 //
2379 
2380 ...
```


###  Channel: Receive
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/transport.go:1400
```go
1389 ...
1390 
1391 		// will keep us from holding it indefinitely if the body
1392 		// close is slow for some reason.
1393 		if bodyClosed != nil {
1394 			<-bodyClosed
1395 		}
1396 		return err
1397 	}
1398 
1399 	for {
1400 		select {           // <-------
1401 		case <-cs.respHeaderRecv:
1402 			return handleResponseHeaders()
1403 		case <-cs.abort:
1404 			select {
1405 			case <-cs.respHeaderRecv:
1406 				// If both cs.respHeaderRecv and cs.abort are signaling,
1407 				// pick respHeaderRecv. The server probably wrote the
1408 				// response and immediately reset the stream.
1409 				// golang.org/issue/49645
1410 				return handleResponseHeaders()
1411 
1412 ...
```


## Replay
**Replaying was not run**.

