# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestBindMiddleware
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/utils_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/utils.go:25
```go
14 ...
15 
16 )
17 
18 // BindKey indicates a default bind key.
19 const BindKey = "_gin-gonic/gin/bindkey"
20 
21 // Bind is a helper function for given interface object and returns a Gin middleware.
22 func Bind(val any) HandlerFunc {
23 	value := reflect.ValueOf(val)
24 	if value.Kind() == reflect.Ptr {
25 		panic(`Bind struct can not be a pointer. Example:           // <-------
26 	Use: gin.Bind(Struct{}) instead of gin.Bind(&Struct{})
27 `)
28 	}
29 	typ := value.Type()
30 
31 	return func(c *Context) {
32 		obj := reflect.New(typ).Interface()
33 		if c.Bind(obj) == nil {
34 			c.Set(BindKey, obj)
35 		}
36 
37 ...
```


## Replay
**Replaying was not run**.

