# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestEvent_CloudEvent_NilOrigin
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/caddy_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/go/pkg/mod/github.com/prometheus/client_golang@v1.19.1/prometheus/registry.go:272
```go
261 ...
262 
263 	descIDs               map[uint64]struct{}
264 	dimHashesByName       map[string]uint64
265 	uncheckedCollectors   []Collector
266 	pedanticChecksEnabled bool
267 }
268 
269 // Register implements Registerer.
270 func (r *Registry) Register(c Collector) error {
271 	var (
272 		descChan           = make(chan *Desc, capDescChan)           // <-------
273 		newDescIDs         = map[uint64]struct{}{}
274 		newDimHashesByName = map[string]uint64{}
275 		collectorID        uint64 // All desc IDs XOR'd together.
276 		duplicateDescErr   error
277 	)
278 	go func() {
279 		c.Describe(descChan)
280 		close(descChan)
281 	}()
282 	r.mtx.Lock()
283 
284 ...
```


## Replay
**Replaying was not run**.

