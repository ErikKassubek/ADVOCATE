# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestTreeDuplicatePath
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree.go:253
```go
242 ...
243 
244 					"'")
245 			}
246 
247 			n.insertChild(path, fullPath, handlers)
248 			return
249 		}
250 
251 		// Otherwise add handle to current node
252 		if n.handlers != nil {
253 			panic("handlers are already registered for path '" + fullPath + "'")           // <-------
254 		}
255 		n.handlers = handlers
256 		n.fullPath = fullPath
257 		return
258 	}
259 }
260 
261 // Search for a wildcard segment and check the name for invalid characters.
262 // Returns -1 as index, if no wildcard was found.
263 func findWildcard(path string) (wildcard string, i int, valid bool) {
264 
265 ...
```


## Replay
**Replaying was not run**.

