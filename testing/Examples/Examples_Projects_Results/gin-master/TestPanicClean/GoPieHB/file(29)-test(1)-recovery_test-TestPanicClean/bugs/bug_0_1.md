# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestPanicClean
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/recovery_test.go:25
```go
14 ...
15 
16 )
17 
18 func TestPanicClean(t *testing.T) {
19 	buffer := new(strings.Builder)
20 	router := New()
21 	password := "my-super-secret-password"
22 	router.Use(RecoveryWithWriter(buffer))
23 	router.GET("/recovery", func(c *Context) {
24 		c.AbortWithStatus(http.StatusBadRequest)
25 		panic("Oupps, Houston, we have a problem")           // <-------
26 	})
27 	// RUN
28 	w := PerformRequest(router, http.MethodGet, "/recovery",
29 		header{
30 			Key:   "Host",
31 			Value: "www.google.com",
32 		},
33 		header{
34 			Key:   "Authorization",
35 			Value: "Bearer " + password,
36 
37 ...
```


## Replay
**Replaying was not run**.

