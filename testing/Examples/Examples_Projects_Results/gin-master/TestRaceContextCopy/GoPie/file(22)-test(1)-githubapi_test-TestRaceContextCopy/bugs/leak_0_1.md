# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRaceContextCopy
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/githubapi_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context.go:281
```go
270 ...
271 
272 		c.Keys = make(map[any]any)
273 	}
274 
275 	c.Keys[key] = value
276 }
277 
278 // Get returns the value for the given key, ie: (value, true).
279 // If the value does not exist it returns (nil, false)
280 func (c *Context) Get(key any) (value any, exists bool) {
281 	c.mu.RLock()           // <-------
282 	defer c.mu.RUnlock()
283 	value, exists = c.Keys[key]
284 	return
285 }
286 
287 // MustGet returns the value for the given key if it exists, otherwise it panics.
288 func (c *Context) MustGet(key any) any {
289 	if value, exists := c.Get(key); exists {
290 		return value
291 	}
292 
293 ...
```


## Replay
**Replaying was not run**.

