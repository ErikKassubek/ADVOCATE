# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestChooseData
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/utils_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/utils.go:100
```go
89 ...
90 
91 }
92 
93 func chooseData(custom, wildcard any) any {
94 	if custom != nil {
95 		return custom
96 	}
97 	if wildcard != nil {
98 		return wildcard
99 	}
100 	panic("negotiation config is invalid")           // <-------
101 }
102 
103 func parseAccept(acceptHeader string) []string {
104 	parts := strings.Split(acceptHeader, ",")
105 	out := make([]string, 0, len(parts))
106 	for _, part := range parts {
107 		if i := strings.IndexByte(part, ';'); i > 0 {
108 			part = part[:i]
109 		}
110 		if part = strings.TrimSpace(part); part != "" {
111 
112 ...
```


## Replay
**Replaying was not run**.

