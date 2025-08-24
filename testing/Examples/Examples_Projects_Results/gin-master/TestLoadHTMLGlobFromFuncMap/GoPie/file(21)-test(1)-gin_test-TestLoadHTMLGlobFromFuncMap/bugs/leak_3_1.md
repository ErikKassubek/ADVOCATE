# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestLoadHTMLGlobFromFuncMap
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/mode.go:69
```go
58 ...
59 
60 		if flag.Lookup("test.v") != nil {
61 			value = TestMode
62 		} else {
63 			value = DebugMode
64 		}
65 	}
66 
67 	switch value {
68 	case DebugMode, "":
69 		atomic.StoreInt32(&ginMode, debugCode)           // <-------
70 	case ReleaseMode:
71 		atomic.StoreInt32(&ginMode, releaseCode)
72 	case TestMode:
73 		atomic.StoreInt32(&ginMode, testCode)
74 	default:
75 		panic("gin mode unknown: " + value + " (available mode: debug release test)")
76 	}
77 	modeName.Store(value)
78 }
79 
80 
81 ...
```


## Replay
**Replaying was not run**.

