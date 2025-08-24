# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestAddRouteFails
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/utils.go:80
```go
69 ...
70 
71 			return err
72 		}
73 	}
74 
75 	return e.EncodeToken(xml.EndElement{Name: start.Name})
76 }
77 
78 func assert1(guard bool, text string) {
79 	if !guard {
80 		panic(text)           // <-------
81 	}
82 }
83 
84 func filterFlags(content string) string {
85 	for i, char := range content {
86 		if char == ' ' || char == ';' {
87 			return content[:i]
88 		}
89 	}
90 	return content
91 
92 ...
```


## Replay
**Replaying was not run**.

