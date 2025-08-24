# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestL05_LeakOnNilChan
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/ADVOCATE/go-patch/src/runtime/panic.go:262
```go
251 ...
252 
253 func panicfloat() {
254 	panicCheck2("floating point error")
255 	panic(floatError)
256 }
257 
258 var memoryError = error(errorString("invalid memory address or nil pointer dereference"))
259 
260 func panicmem() {
261 	panicCheck2("invalid memory address or nil pointer dereference")
262 	panic(memoryError)           // <-------
263 }
264 
265 func panicmemAddr(addr uintptr) {
266 	panicCheck2("invalid memory address or nil pointer dereference")
267 	panic(errorAddressString{msg: "invalid memory address or nil pointer dereference", addr: addr})
268 }
269 
270 // Create a new deferred function fn, which has no arguments and results.
271 // The compiler turns a defer statement into a call to this.
272 func deferproc(fn func()) {
273 
274 ...
```


## Replay
**Replaying was not run**.

