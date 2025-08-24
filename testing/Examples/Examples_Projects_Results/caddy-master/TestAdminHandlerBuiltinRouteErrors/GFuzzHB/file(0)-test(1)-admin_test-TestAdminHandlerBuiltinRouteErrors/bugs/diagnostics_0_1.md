# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestAdminHandlerBuiltinRouteErrors
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/admin_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/go/pkg/mod/github.com/prometheus/client_golang@v1.19.1/prometheus/registry.go:375
```go
364 ...
365 
366 	var (
367 		descChan    = make(chan *Desc, capDescChan)
368 		descIDs     = map[uint64]struct{}{}
369 		collectorID uint64 // All desc IDs XOR'd together.
370 	)
371 	go func() {
372 		c.Describe(descChan)
373 		close(descChan)
374 	}()
375 	for desc := range descChan {           // <-------
376 		if _, exists := descIDs[desc.id]; !exists {
377 			collectorID ^= desc.id
378 			descIDs[desc.id] = struct{}{}
379 		}
380 	}
381 
382 	r.mtx.RLock()
383 	if _, exists := r.collectorsByID[collectorID]; !exists {
384 		r.mtx.RUnlock()
385 		return false
386 
387 ...
```


###  Channel: Close
-> /home/ilian/Projects/go/pkg/mod/github.com/prometheus/client_golang@v1.19.1/prometheus/registry.go:373
```go
362 ...
363 
364 // Unregister implements Registerer.
365 func (r *Registry) Unregister(c Collector) bool {
366 	var (
367 		descChan    = make(chan *Desc, capDescChan)
368 		descIDs     = map[uint64]struct{}{}
369 		collectorID uint64 // All desc IDs XOR'd together.
370 	)
371 	go func() {
372 		c.Describe(descChan)
373 		close(descChan)           // <-------
374 	}()
375 	for desc := range descChan {
376 		if _, exists := descIDs[desc.id]; !exists {
377 			collectorID ^= desc.id
378 			descIDs[desc.id] = struct{}{}
379 		}
380 	}
381 
382 	r.mtx.RLock()
383 	if _, exists := r.collectorsByID[collectorID]; !exists {
384 
385 ...
```


## Replay
**Replaying was not run**.

