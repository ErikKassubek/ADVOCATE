# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestTreeWildcardConflict
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree.go:240
```go
229 ...
230 
231 					continue walk
232 				}
233 
234 				// Wildcard conflict
235 				pathSeg := path
236 				if n.nType != catchAll {
237 					pathSeg, _, _ = strings.Cut(pathSeg, "/")
238 				}
239 				prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
240 				panic("'" + pathSeg +           // <-------
241 					"' in new path '" + fullPath +
242 					"' conflicts with existing wildcard '" + n.path +
243 					"' in existing prefix '" + prefix +
244 					"'")
245 			}
246 
247 			n.insertChild(path, fullPath, handlers)
248 			return
249 		}
250 
251 
252 ...
```


## Replay
**Replaying was not run**.

