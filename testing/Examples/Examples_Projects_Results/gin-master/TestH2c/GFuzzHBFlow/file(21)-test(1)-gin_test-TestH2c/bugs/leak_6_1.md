# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestH2c
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/server.go:444
```go
433 ...
434 
435 	sc := &serverConn{
436 		srv:                         s,
437 		hs:                          http1srv,
438 		conn:                        c,
439 		baseCtx:                     baseCtx,
440 		remoteAddrStr:               c.RemoteAddr().String(),
441 		bw:                          newBufferedWriter(s.group, c, conf.WriteByteTimeout),
442 		handler:                     opts.handler(),
443 		streams:                     make(map[uint32]*stream),
444 		readFrameCh:                 make(chan readFrameResult),           // <-------
445 		wantWriteFrameCh:            make(chan FrameWriteRequest, 8),
446 		serveMsgCh:                  make(chan interface{}, 8),
447 		wroteFrameCh:                make(chan frameWriteResult, 1), // buffered; one send in writeFrameAsync
448 		bodyReadCh:                  make(chan bodyReadMsg),         // buffering doesn't matter either way
449 		doneServing:                 make(chan struct{}),
450 		clientMaxStreams:            math.MaxUint32, // Section 6.5.2: "Initially, there is no limit to this value"
451 		advMaxStreams:               conf.MaxConcurrentStreams,
452 		initialStreamSendWindowSize: initialWindowSize,
453 		initialStreamRecvWindowSize: conf.MaxUploadBufferPerStream,
454 		maxFrameSize:                initialMaxFrameSize,
455 
456 ...
```


## Replay
**Replaying was not run**.

