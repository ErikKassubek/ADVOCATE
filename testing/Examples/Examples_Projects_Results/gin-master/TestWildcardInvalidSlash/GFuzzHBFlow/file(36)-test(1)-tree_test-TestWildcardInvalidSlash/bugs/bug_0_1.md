# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestWildcardInvalidSlash
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree.go:373
```go
362 ...
363 
364 				"' in new path '" + fullPath +
365 				"' conflicts with existing path segment '" + pathSeg +
366 				"' in existing prefix '" + n.path + pathSeg +
367 				"'")
368 		}
369 
370 		// currently fixed width 1 for '/'
371 		i--
372 		if i < 0 || path[i] != '/' {
373 			panic("no / before catch-all in path '" + fullPath + "'")           // <-------
374 		}
375 
376 		n.path = path[:i]
377 
378 		// First node: catchAll node with empty path
379 		child := &node{
380 			wildChild: true,
381 			nType:     catchAll,
382 			fullPath:  fullPath,
383 		}
384 
385 ...
```


## Replay
**Replaying was not run**.

