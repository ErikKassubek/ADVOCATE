# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestPanicWithAbort
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go:83
```go
72 ...
73 
74 	SetMode(TestMode)
75 }
76 
77 // TestPanicWithAbort assert that panic has been recovered even if context.Abort was used.
78 func TestPanicWithAbort(t *testing.T) {
79 	router := New()
80 	router.Use(RecoveryWithWriter(nil))
81 	router.GET("/recovery", func(c *Context) {
82 		c.AbortWithStatus(http.StatusBadRequest)
83 		panic("Oupps, Houston, we have a problem")           // <-------
84 	})
85 	// RUN
86 	w := PerformRequest(router, http.MethodGet, "/recovery")
87 	// TEST
88 	assert.Equal(t, http.StatusBadRequest, w.Code)
89 }
90 
91 func TestMaskAuthorization(t *testing.T) {
92 	secret := "Bearer aaaabbbbccccddddeeeeffff"
93 	headers := []string{
94 
95 ...
```


## Replay
**Replaying was not run**.

