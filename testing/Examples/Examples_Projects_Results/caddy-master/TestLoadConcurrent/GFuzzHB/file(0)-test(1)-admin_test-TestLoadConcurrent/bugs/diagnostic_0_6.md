# Diagnostic: Possible Receive on Closed Channel

The analyzer detected a possible receive on a closed channel.
Although the receive on a closed channel did not occur during the recording, it is possible that it will occur, based on the happens before relation.This is not necessarily a bug, but it can be an indication of a bug.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestLoadConcurrent
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/admin_test.go

## Bug Elements
The elements involved in the found diagnostic are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/go/pkg/mod/github.com/prometheus/client_golang@v1.19.1/prometheus/registry.go:290
```go
279 ...
280 
281 	}()
282 	r.mtx.Lock()
283 	defer func() {
284 		// Drain channel in case of premature return to not leak a goroutine.
285 		for range descChan {
286 		}
287 		r.mtx.Unlock()
288 	}()
289 	// Conduct various tests...
290 	for desc := range descChan {           // <-------
291 
292 		// Is the descriptor valid at all?
293 		if desc.err != nil {
294 			return fmt.Errorf("descriptor %s is invalid: %w", desc, desc.err)
295 		}
296 
297 		// Is the descID unique?
298 		// (In other words: Is the fqName + constLabel combination unique?)
299 		if _, exists := r.descIDs[desc.id]; exists {
300 			duplicateDescErr = fmt.Errorf("descriptor %s already exists with the same fully-qualified name and const label values", desc)
301 
302 ...
```


###  Channel: Close
-> /home/ilian/Projects/go/pkg/mod/github.com/prometheus/client_golang@v1.19.1/prometheus/registry.go:280
```go
269 ...
270 
271 	var (
272 		descChan           = make(chan *Desc, capDescChan)
273 		newDescIDs         = map[uint64]struct{}{}
274 		newDimHashesByName = map[string]uint64{}
275 		collectorID        uint64 // All desc IDs XOR'd together.
276 		duplicateDescErr   error
277 	)
278 	go func() {
279 		c.Describe(descChan)
280 		close(descChan)           // <-------
281 	}()
282 	r.mtx.Lock()
283 	defer func() {
284 		// Drain channel in case of premature return to not leak a goroutine.
285 		for range descChan {
286 		}
287 		r.mtx.Unlock()
288 	}()
289 	// Conduct various tests...
290 	for desc := range descChan {
291 
292 ...
```


## Replay
The bug is a potential bug.
The analyzer has tries to rewrite the trace in such a way, that the bug will be triggered when replaying the trace.

**Replaying ended without confirming the bug**.

It exited with the following code: 0

The replay finished without being able to confirm the predicted bug. If the given trace was a directly recorded trace, this is the expected behavior. If it was rewritten by the analyzer, this could be an indication that something went wrong during rewrite or replay.

