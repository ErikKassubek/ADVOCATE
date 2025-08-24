# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRecoveryWithWriterWithCustomRecovery
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go:243
```go
232 ...
233 
234 	buffer := new(strings.Builder)
235 	router := New()
236 	DefaultErrorWriter = buffer
237 	handleRecovery := func(c *Context, err any) {
238 		errBuffer.WriteString(err.(string))
239 		c.AbortWithStatus(http.StatusBadRequest)
240 	}
241 	router.Use(RecoveryWithWriter(DefaultErrorWriter, handleRecovery))
242 	router.GET("/recovery", func(_ *Context) {
243 		panic("Oupps, Houston, we have a problem")           // <-------
244 	})
245 	// RUN
246 	w := PerformRequest(router, http.MethodGet, "/recovery")
247 	// TEST
248 	assert.Equal(t, http.StatusBadRequest, w.Code)
249 	assert.Contains(t, buffer.String(), "panic recovered")
250 	assert.Contains(t, buffer.String(), "Oupps, Houston, we have a problem")
251 	assert.Contains(t, buffer.String(), t.Name())
252 	assert.NotContains(t, buffer.String(), "GET /recovery")
253 
254 
255 ...
```


## Replay
**Replaying was not run**.

