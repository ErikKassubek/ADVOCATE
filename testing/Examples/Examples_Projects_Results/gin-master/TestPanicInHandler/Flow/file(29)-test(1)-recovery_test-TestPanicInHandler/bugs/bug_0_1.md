# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestPanicInHandler
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go:55
```go
44 ...
45 
46 	assert.NotContains(t, buffer.String(), password)
47 }
48 
49 // TestPanicInHandler assert that panic has been recovered.
50 func TestPanicInHandler(t *testing.T) {
51 	buffer := new(strings.Builder)
52 	router := New()
53 	router.Use(RecoveryWithWriter(buffer))
54 	router.GET("/recovery", func(_ *Context) {
55 		panic("Oupps, Houston, we have a problem")           // <-------
56 	})
57 	// RUN
58 	w := PerformRequest(router, http.MethodGet, "/recovery")
59 	// TEST
60 	assert.Equal(t, http.StatusInternalServerError, w.Code)
61 	assert.Contains(t, buffer.String(), "panic recovered")
62 	assert.Contains(t, buffer.String(), "Oupps, Houston, we have a problem")
63 	assert.Contains(t, buffer.String(), t.Name())
64 	assert.NotContains(t, buffer.String(), "GET /recovery")
65 
66 
67 ...
```


## Replay
**Replaying was not run**.

