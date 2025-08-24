# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestContextSetGet
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context.go:292
```go
281 ...
282 
283 	value, exists = c.Keys[key]
284 	return
285 }
286 
287 // MustGet returns the value for the given key if it exists, otherwise it panics.
288 func (c *Context) MustGet(key any) any {
289 	if value, exists := c.Get(key); exists {
290 		return value
291 	}
292 	panic(fmt.Sprintf("key %v does not exist", key))           // <-------
293 }
294 
295 func getTyped[T any](c *Context, key any) (res T) {
296 	if val, ok := c.Get(key); ok && val != nil {
297 		res, _ = val.(T)
298 	}
299 	return
300 }
301 
302 // GetString returns the value associated with the key as a string.
303 
304 ...
```


## Replay
**Replaying was not run**.

