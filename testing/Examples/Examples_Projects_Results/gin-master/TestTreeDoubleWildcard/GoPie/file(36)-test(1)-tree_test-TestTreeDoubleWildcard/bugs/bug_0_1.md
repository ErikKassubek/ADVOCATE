# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestTreeDoubleWildcard
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree.go:308
```go
297 ...
298 
299 	for {
300 		// Find prefix until first wildcard
301 		wildcard, i, valid := findWildcard(path)
302 		if i < 0 { // No wildcard found
303 			break
304 		}
305 
306 		// The wildcard name must only contain one ':' or '*' character
307 		if !valid {
308 			panic("only one wildcard per path segment is allowed, has: '" +           // <-------
309 				wildcard + "' in path '" + fullPath + "'")
310 		}
311 
312 		// check if the wildcard has a name
313 		if len(wildcard) < 2 {
314 			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
315 		}
316 
317 		if wildcard[0] == ':' { // param
318 			if i > 0 {
319 
320 ...
```


## Replay
**Replaying was not run**.

