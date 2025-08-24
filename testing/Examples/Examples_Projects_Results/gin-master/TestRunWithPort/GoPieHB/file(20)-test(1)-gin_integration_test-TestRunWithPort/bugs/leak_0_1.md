# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRunWithPort
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_integration_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/debug.go:21
```go
10 ...
11 
12 	"strings"
13 	"sync/atomic"
14 )
15 
16 const ginSupportMinGoVer = 21
17 
18 // IsDebugging returns true if the framework is running in debug mode.
19 // Use SetMode(gin.ReleaseMode) to disable debug mode.
20 func IsDebugging() bool {
21 	return atomic.LoadInt32(&ginMode) == debugCode           // <-------
22 }
23 
24 // DebugPrintRouteFunc indicates debug log output format.
25 var DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)
26 
27 // DebugPrintFunc indicates debug log output format.
28 var DebugPrintFunc func(format string, values ...any)
29 
30 func debugPrintRoute(httpMethod, absolutePath string, handlers HandlersChain) {
31 	if IsDebugging() {
32 
33 ...
```


## Replay
**Replaying was not run**.

