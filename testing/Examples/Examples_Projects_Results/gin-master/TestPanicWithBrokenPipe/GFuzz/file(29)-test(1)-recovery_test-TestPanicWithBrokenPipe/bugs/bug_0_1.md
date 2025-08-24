# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestPanicWithBrokenPipe
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go:152
```go
141 ...
142 
143 			router := New()
144 			router.Use(RecoveryWithWriter(&buf))
145 			router.GET("/recovery", func(c *Context) {
146 				// Start writing response
147 				c.Header("X-Test", "Value")
148 				c.Status(expectCode)
149 
150 				// Oops. Client connection closed
151 				e := &net.OpError{Err: &os.SyscallError{Err: errno}}
152 				panic(e)           // <-------
153 			})
154 			// RUN
155 			w := PerformRequest(router, http.MethodGet, "/recovery")
156 			// TEST
157 			assert.Equal(t, expectCode, w.Code)
158 			assert.Contains(t, strings.ToLower(buf.String()), expectMsg)
159 		})
160 	}
161 }
162 
163 
164 ...
```


## Replay
**Replaying was not run**.

