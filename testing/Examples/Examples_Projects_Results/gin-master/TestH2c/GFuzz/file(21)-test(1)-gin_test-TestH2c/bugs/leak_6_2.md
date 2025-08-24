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
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/transport.go:826
```go
815 ...
816 
817 func (t *Transport) NewClientConn(c net.Conn) (*ClientConn, error) {
818 	return t.newClientConn(c, t.disableKeepAlives())
819 }
820 
821 func (t *Transport) newClientConn(c net.Conn, singleUse bool) (*ClientConn, error) {
822 	conf := configFromTransport(t)
823 	cc := &ClientConn{
824 		t:                           t,
825 		tconn:                       c,
826 		readerDone:                  make(chan struct{}),           // <-------
827 		nextStreamID:                1,
828 		maxFrameSize:                16 << 10, // spec default
829 		initialWindowSize:           65535,    // spec default
830 		initialStreamRecvWindowSize: conf.MaxUploadBufferPerStream,
831 		maxConcurrentStreams:        initialMaxConcurrentStreams, // "infinite", per spec. Use a smaller value until we have received server settings.
832 		peerMaxHeaderListSize:       0xffffffffffffffff,          // "infinite", per spec. Use 2^64-1 instead.
833 		streams:                     make(map[uint32]*clientStream),
834 		singleUse:                   singleUse,
835 		seenSettingsChan:            make(chan struct{}),
836 		wantSettingsAck:             true,
837 
838 ...
```


## Replay
**Replaying was not run**.

