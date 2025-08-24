# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestEmptyWildcardName
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree.go:314
```go
303 ...
304 
305 
306 		// The wildcard name must only contain one ':' or '*' character
307 		if !valid {
308 			panic("only one wildcard per path segment is allowed, has: '" +
309 				wildcard + "' in path '" + fullPath + "'")
310 		}
311 
312 		// check if the wildcard has a name
313 		if len(wildcard) < 2 {
314 			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")           // <-------
315 		}
316 
317 		if wildcard[0] == ':' { // param
318 			if i > 0 {
319 				// Insert prefix before the current wildcard
320 				n.path = path[:i]
321 				path = path[i:]
322 			}
323 
324 			child := &node{
325 
326 ...
```


## Replay
**Replaying was not run**.

