# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestLastChar
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/utils_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/utils.go:119
```go
108 ...
109 
110 		if part = strings.TrimSpace(part); part != "" {
111 			out = append(out, part)
112 		}
113 	}
114 	return out
115 }
116 
117 func lastChar(str string) uint8 {
118 	if str == "" {
119 		panic("The length of the string can't be 0")           // <-------
120 	}
121 	return str[len(str)-1]
122 }
123 
124 func nameOfFunction(f any) string {
125 	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
126 }
127 
128 func joinPaths(absolutePath, relativePath string) string {
129 	if relativePath == "" {
130 
131 ...
```


## Replay
**Replaying was not run**.

