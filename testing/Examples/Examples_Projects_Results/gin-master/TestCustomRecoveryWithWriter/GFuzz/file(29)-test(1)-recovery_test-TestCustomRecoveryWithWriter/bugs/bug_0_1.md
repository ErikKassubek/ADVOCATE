# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCustomRecoveryWithWriter
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go:173
```go
162 ...
163 
164 	errBuffer := new(strings.Builder)
165 	buffer := new(strings.Builder)
166 	router := New()
167 	handleRecovery := func(c *Context, err any) {
168 		errBuffer.WriteString(err.(string))
169 		c.AbortWithStatus(http.StatusBadRequest)
170 	}
171 	router.Use(CustomRecoveryWithWriter(buffer, handleRecovery))
172 	router.GET("/recovery", func(_ *Context) {
173 		panic("Oupps, Houston, we have a problem")           // <-------
174 	})
175 	// RUN
176 	w := PerformRequest(router, http.MethodGet, "/recovery")
177 	// TEST
178 	assert.Equal(t, http.StatusBadRequest, w.Code)
179 	assert.Contains(t, buffer.String(), "panic recovered")
180 	assert.Contains(t, buffer.String(), "Oupps, Houston, we have a problem")
181 	assert.Contains(t, buffer.String(), t.Name())
182 	assert.NotContains(t, buffer.String(), "GET /recovery")
183 
184 
185 ...
```


## Replay
**Replaying was not run**.

