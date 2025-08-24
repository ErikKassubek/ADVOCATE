# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestTreeCatchAllConflictRoot
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree.go:363
```go
352 ...
353 
354 		if i+len(wildcard) != len(path) {
355 			panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
356 		}
357 
358 		if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
359 			pathSeg := ""
360 			if len(n.children) != 0 {
361 				pathSeg, _, _ = strings.Cut(n.children[0].path, "/")
362 			}
363 			panic("catch-all wildcard '" + path +           // <-------
364 				"' in new path '" + fullPath +
365 				"' conflicts with existing path segment '" + pathSeg +
366 				"' in existing prefix '" + n.path + pathSeg +
367 				"'")
368 		}
369 
370 		// currently fixed width 1 for '/'
371 		i--
372 		if i < 0 || path[i] != '/' {
373 			panic("no / before catch-all in path '" + fullPath + "'")
374 
375 ...
```


## Replay
**Replaying was not run**.

