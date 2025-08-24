# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRunTooMuchParams
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_integration_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/utils.go:152
```go
141 ...
142 
143 		if port := os.Getenv("PORT"); port != "" {
144 			debugPrint("Environment variable PORT=\"%s\"", port)
145 			return ":" + port
146 		}
147 		debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
148 		return ":8080"
149 	case 1:
150 		return addr[0]
151 	default:
152 		panic("too many parameters")           // <-------
153 	}
154 }
155 
156 // https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
157 func isASCII(s string) bool {
158 	for i := 0; i < len(s); i++ {
159 		if s[i] > unicode.MaxASCII {
160 			return false
161 		}
162 	}
163 
164 ...
```


## Replay
**Replaying was not run**.

