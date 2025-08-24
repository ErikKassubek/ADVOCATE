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
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/server.go:1160
```go
1149 ...
1150 
1151 			if VerboseLogs {
1152 				sc.vlogf("http2: server: client %v said hello", sc.conn.RemoteAddr())
1153 			}
1154 		}
1155 		return err
1156 	}
1157 }
1158 
1159 var errChanPool = sync.Pool{
1160 	New: func() interface{} { return make(chan error, 1) },           // <-------
1161 }
1162 
1163 var writeDataPool = sync.Pool{
1164 	New: func() interface{} { return new(writeData) },
1165 }
1166 
1167 // writeDataFromHandler writes DATA response frames from a handler on
1168 // the given stream.
1169 func (sc *serverConn) writeDataFromHandler(stream *stream, data []byte, endStream bool) error {
1170 	ch := errChanPool.Get().(chan error)
1171 
1172 ...
```


## Replay
**Replaying was not run**.

