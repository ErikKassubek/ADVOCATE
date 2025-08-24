# Leak: Leak on sync.Mutex

The analyzer detected a leak on a sync.Mutex.
A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.
A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRaceContextCopy
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/githubapi_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Mutex: Lock
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


###  Mutex: RLock
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

