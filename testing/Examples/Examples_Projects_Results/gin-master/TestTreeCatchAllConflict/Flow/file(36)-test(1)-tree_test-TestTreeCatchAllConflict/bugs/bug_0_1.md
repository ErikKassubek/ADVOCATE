# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestTreeCatchAllConflict
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree.go:355
```go
344 ...
345 
346 			}
347 
348 			// Otherwise we're done. Insert the handle in the new leaf
349 			n.handlers = handlers
350 			return
351 		}
352 
353 		// catchAll
354 		if i+len(wildcard) != len(path) {
355 			panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")           // <-------
356 		}
357 
358 		if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
359 			pathSeg := ""
360 			if len(n.children) != 0 {
361 				pathSeg, _, _ = strings.Cut(n.children[0].path, "/")
362 			}
363 			panic("catch-all wildcard '" + path +
364 				"' in new path '" + fullPath +
365 				"' conflicts with existing path segment '" + pathSeg +
366 
367 ...
```


## Replay
**Replaying was not run**.

