# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestEscapedColon
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_integration_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree.go:272
```go
261 ...
262 
263 func findWildcard(path string) (wildcard string, i int, valid bool) {
264 	// Find start
265 	escapeColon := false
266 	for start, c := range []byte(path) {
267 		if escapeColon {
268 			escapeColon = false
269 			if c == ':' {
270 				continue
271 			}
272 			panic("invalid escape string in path '" + path + "'")           // <-------
273 		}
274 		if c == '\\' {
275 			escapeColon = true
276 			continue
277 		}
278 		// A wildcard starts with ':' (param) or '*' (catch-all)
279 		if c != ':' && c != '*' {
280 			continue
281 		}
282 
283 
284 ...
```


## Replay
**Replaying was not run**.

