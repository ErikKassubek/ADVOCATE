# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestContextSetGetPanicsWhenKeyNotComparable
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/ADVOCATE/go-patch/src/runtime/alg.go:171
```go
160 ...
161 
162 //go:linkname nilinterhash
163 func nilinterhash(p unsafe.Pointer, h uintptr) uintptr {
164 	a := (*eface)(p)
165 	t := a._type
166 	if t == nil {
167 		return h
168 	}
169 	if t.Equal == nil {
170 		// See comment in interhash above.
171 		panic(errorString("hash of unhashable type " + toRType(t).string()))           // <-------
172 	}
173 	if isDirectIface(t) {
174 		return c1 * typehash(t, unsafe.Pointer(&a.data), h^c0)
175 	} else {
176 		return c1 * typehash(t, a.data, h^c0)
177 	}
178 }
179 
180 // typehash computes the hash of the object of type t at address p.
181 // h is the seed.
182 
183 ...
```


## Replay
**Replaying was not run**.

