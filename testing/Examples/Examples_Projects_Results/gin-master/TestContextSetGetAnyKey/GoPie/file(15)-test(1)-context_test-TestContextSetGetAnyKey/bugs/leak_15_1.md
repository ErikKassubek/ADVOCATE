# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestContextSetGetAnyKey
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context.go:269
```go
258 ...
259 
260 }
261 
262 /************************************/
263 /******** METADATA MANAGEMENT********/
264 /************************************/
265 
266 // Set is used to store a new key/value pair exclusively for this context.
267 // It also lazy initializes  c.Keys if it was not used previously.
268 func (c *Context) Set(key any, value any) {
269 	c.mu.Lock()           // <-------
270 	defer c.mu.Unlock()
271 	if c.Keys == nil {
272 		c.Keys = make(map[any]any)
273 	}
274 
275 	c.Keys[key] = value
276 }
277 
278 // Get returns the value for the given key, ie: (value, true).
279 // If the value does not exist it returns (nil, false)
280 
281 ...
```


## Replay
**Replaying was not run**.

